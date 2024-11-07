package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Sph3ricalPeter/frbench/external"
)

const (
	AnthClaude35Haiku = "claude-3-5-haiku-20241022"
	AnthClaude3Haiku  = "claude-3-haiku-20240307"
)

func main() {
	// read cache flag from args
	useCache := flag.Bool("c", false, "use cache")
	flag.Parse()

	appliedPatches := make([][]byte, 0)

	err := runTests(1)
	if err == nil {
		panic("tests already pass, no need to apply patches ...")
	}

	// TODO: i need to incrementally add the context from the model's response as well as the additional requirements ...
	// to kind of mimic the chat experience
	// TODO: also add caching for when the same request gets sent multiple times
	// so that when i'm testing step 2, i don't keep re-sending the same step 1 prompt
	promptPayload := external.AnthMessagesPayload{
		Model:     AnthClaude3Haiku,
		MaxTokens: 2048,
		System:    external.SystemPrompt,
		Messages:  make([]external.AnthMessage, 0),
	}

	score := 0
	for i := 1; i < 4; i++ {
		fmt.Printf("Patching #%d ...\n", i)

		promptBytes, err := os.ReadFile(fmt.Sprintf("tests/%d_test.go", i))
		if err != nil {
			panic(fmt.Errorf("error reading prompt message content from file: %w", err))
		}

		promptPayload.Messages = append(promptPayload.Messages, external.AnthMessage{
			Role:    "user",
			Content: string(promptBytes),
		})

		promptFile, err := os.Create(fmt.Sprintf("data/anth-prompt-%d.json", i))
		if err != nil {
			panic(fmt.Errorf("error creating prompt file: %w", err))
		}
		defer promptFile.Close()
		err = json.NewEncoder(promptFile).Encode(promptPayload)
		if err != nil {
			panic(fmt.Errorf("error writing prompt to file: %w", err))
		}

		var respBytes []byte
		wasCached := false
		if _, err := os.Stat(fmt.Sprintf("cache/%d.json", i)); err == nil && *useCache {
			fmt.Println("Using cache ...")
			respBytes, err = os.ReadFile(fmt.Sprintf("cache/%d.json", i))
			if err != nil {
				panic(fmt.Errorf("error reading cache file: %w", err))
			}
			wasCached = true
		} else {
			fmt.Println("Sending prompt ...")

			respBytes, err = external.SendPrompt(promptPayload)
			if err != nil {
				panic(fmt.Errorf("error sending prompt: %w", err))
			}
		}

		resp, err := external.ParseResponse(respBytes)
		if err != nil {
			panic(fmt.Errorf("error parsing response: %w", err))
		}

		promptPayload.Messages = append(promptPayload.Messages, external.AnthMessage{
			Role:    "assistant",
			Content: resp.Content[0].Text,
		})

		patchFpath := fmt.Sprintf("data/anth-patch-%d.patch", i)
		err = external.CreatePatchFile(*resp, patchFpath)
		if err != nil {
			panic(fmt.Errorf("error creating patch file: %w", err))
		}
		patchBytes, err := os.ReadFile(patchFpath)
		if err != nil {
			panic(fmt.Errorf("error reading patch file: %w", err))
		}

		// applying the patch given by the model (& parsed) should make the tests pass
		err = applyPatch(patchBytes)
		if err != nil {
			fmt.Printf("error applying patch: %s\n", err.Error())
			break
		} else {
			appliedPatches = append(appliedPatches, patchBytes)
		}

		// copy test file from tests/ to app/
		err = runCommand(fmt.Sprintf("cp tests/%d_test.go app/", i))
		if err != nil {
			panic(fmt.Errorf("error copying test file: %w", err))
		}
		err = runTests(i)
		if err != nil {
			fmt.Println("Patch FAILED! ❌")
		} else {
			fmt.Println("Patch OK! ✅")
			score += 1
		}

		// cache only the requests which resulted in the patch being applied successfully
		// TODO: maybe use hash of the prompt content as the cache key instead of the index
		if err == nil && !wasCached {
			cacheFile, err := os.Create(fmt.Sprintf("cache/%d.json", i))
			if err != nil {
				panic(fmt.Errorf("error creating cache file: %w", err))
			}
			defer cacheFile.Close()
			_, err = cacheFile.Write(respBytes)
			if err != nil {
				panic(fmt.Errorf("error writing to cache file: %w", err))
			}
		}

		if !wasCached {
			fmt.Printf("Used %d input tokens, %d output tokens\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)
		}

		// wait for input to revert patches
		fmt.Println("Press ENTER to do next patch or revert if it's the last one ...")
		_, _ = os.Stdin.Read(make([]byte, 1))

		if err != nil {
			break
		}
	}

	fmt.Printf("All Done! Score: %d/%d\n", score, 2)

	for i := len(appliedPatches) - 1; i >= 0; i-- {
		revertPatch(appliedPatches[i])
		runCommand(fmt.Sprintf("rm app/%d_test.go", i+1))
	}
}

func applyPatch(patch []byte) error {
	return runCommandWithInput("patch -ruN -d app/", bytes.NewReader(patch))
}

func revertPatch(patch []byte) error {
	return runCommandWithInput("patch -ruN --reverse -d app/", bytes.NewReader(patch))
}

func runTests(i int) error {
	return runCommand(fmt.Sprintf("go test -v -run \".*_%d\" ./app/", i))
}

func runCommand(cmd string) error {
	return runCommandWithInput(cmd, os.Stdin)
}

func runCommandWithInput(cmd string, input io.Reader) error {
	parts := strings.Split(cmd, " ")

	c := exec.Command(parts[0], parts[1:]...)
	c.Stdin = input
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	fmt.Println(c.String())

	return c.Run()
}
