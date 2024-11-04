package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/Sph3ricalPeter/frbench/external"
)

func main() {
	appliedPatches := make([][]byte, 0)

	err := runTests()
	if err == nil {
		panic("tests already pass, no need to apply patches ...")
	}

	for i := 0; i < 1; i++ {
		data, err := external.SendPrompt()
		if err != nil {
			panic(fmt.Errorf("error sending prompt: %w", err))
		}

		resp, err := external.ParseResponse(data)
		if err != nil {
			panic(fmt.Errorf("error parsing response: %w", err))
		}

		patchFpath := fmt.Sprintf("data/anth-patch-%d.patch", i)

		patchFile, err := os.Create(patchFpath)
		if err != nil {
			panic(fmt.Errorf("error creating patch file: %w", err))
		}
		defer patchFile.Close()

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
			panic(fmt.Errorf("error applying patch: %w", err))
		}
		appliedPatches = append(appliedPatches, patchBytes)

		err = runTests()
		if err != nil {
			fmt.Println("Patch FAILED! ❌")
		} else {
			fmt.Println("Patch OK! ✅")
		}
	}

	// wait for input to revert patches
	fmt.Println("Press ENTER to revert patches ...")
	_, _ = os.Stdin.Read(make([]byte, 1))

	for _, patch := range appliedPatches {
		revertPatch(patch)
	}
}

func applyPatch(patch []byte) error {
	return runCommandWithInput("patch -ruN -d app/", bytes.NewReader(patch))
}

func revertPatch(patch []byte) error {
	return runCommandWithInput("patch -ruN --reverse -d app/", bytes.NewReader(patch))
}

func runTests() error {
	return runCommand("go test -v ./app/")
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
