package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Sph3ricalPeter/frbench/common"
	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/external/google"
)

type Benchmark struct {
	con   external.Connector
	score int
}

func main() {
	// read cache flag from args
	useCache := flag.Bool("c", false, "use cache")
	flag.Parse()

	err := runTests(1)
	if err == nil {
		panic("tests already pass, no need to apply patches ...")
	}

	benchmarks := []Benchmark{
		{
			con:   google.NewGoogleConnector(google.Gemini15Flash),
			score: 0,
		},
		// {
		// 	con:   anth.NewAnthConnector(anth.Claude3Haiku),
		// 	score: 0,
		// },
	}

	appliedPatches := map[int][]byte{}
	invalidPatches := map[int]*string{}

	// TODO: somehow do this dynamically based on the provided test files
	// so probably load files based on regex fname matching and it over them
	testCount := 3

	for _, bench := range benchmarks {
		for i := 1; i < testCount+1; i++ {
			fmt.Printf("Running #%d ...\n", i)

			fmt.Println("Sending prompt ...")
			promptBytes, err := os.ReadFile(fmt.Sprintf("tests/%d_test.go", i))
			if err != nil {
				fmt.Printf("error reading test file: %s\n", err.Error())
				break
			}
			result, err := bench.con.SendPrompt(external.SendPromptData{
				Role:     external.RoleUser,
				Prompt:   promptBytes,
				UseCache: *useCache,
				Number:   i,
			})
			if err != nil {
				fmt.Printf("error sending prompt: %s\n", err.Error())
				break
			}
			fmt.Println("OK.")
			if result.CacheKey == nil {
				fmt.Printf("Used %d input / %d output tokens.\n", result.Usage.InputTokens, result.Usage.OutputTokens)
			}

			updatedPatch, err := doPatch([]byte(result.Content), i)
			if err != nil {
				fmt.Printf("error doing patch: %s\n", err.Error())
				invalidPatches[i] = result.CacheKey
				break
			}
			appliedPatches[i] = updatedPatch

			err = verifyPatch(i)
			if err != nil {
				fmt.Println("Patch BAD! ❌")
				invalidPatches[i] = result.CacheKey
				break
			} else {
				fmt.Println("Patch OK! ✅")
				bench.score++
			}

			// wait for input to revert patches
			fmt.Println("Press ENTER to do next patch or revert if it's the last one ...")
			_, _ = os.Stdin.Read(make([]byte, 1))
		}

		fmt.Printf("All Done! %s scored: %d/%d\n", bench.con.GetModelName(), bench.score, testCount)

		for i := len(appliedPatches) - 1; i >= 0; i-- {
			fmt.Printf("Reverting patch #%d ...\n", i+1)
			common.CheckErr(runCommand(fmt.Sprintf("rm app/%d_test.go", i+1)))
			common.CheckErr(revertPatch(appliedPatches[i]))
		}

		// prompt is the key for the cache
		for i, cacheKey := range invalidPatches {
			if cacheKey == nil {
				continue
			}
			fmt.Printf("Removing cache for invalid patch #%d ...\n", i)
			common.CheckErr(bench.con.InvalidateCachedPrompt(*cacheKey))
		}
	}
}

func doPatch(patch []byte, i int) ([]byte, error) {
	patchBytes, err := common.WritePatchFile(patch, fmt.Sprintf("data/patch-%d.patch", i))
	if err != nil {
		return nil, fmt.Errorf("error writing patch file: %w", err)
	}
	err = applyPatch(patchBytes)
	if err != nil {
		return nil, fmt.Errorf("error applying patch: %w", err)
	}
	return patchBytes, nil
}

func verifyPatch(i int) error {
	err := runCommand(fmt.Sprintf("cp tests/%d_test.go app/", i))
	if err != nil {
		return fmt.Errorf("error copying test file: %w", err)
	}
	err = runTests(i)
	if err != nil {
		return fmt.Errorf("patch failed: %w", err)
	}
	return nil
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
