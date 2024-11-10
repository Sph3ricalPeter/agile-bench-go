package internal

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Sph3ricalPeter/frbench/internal/common"
)

func Patch(patch []byte, i int) ([]byte, error) {
	fmt.Printf("Applying patch #%d ...\n", i)
	patchBytes, err := writePatchFile(patch, fmt.Sprintf("data/patch-%d.patch", i))
	fmt.Println(string(patchBytes))
	if err != nil {
		return nil, fmt.Errorf("error writing patch file: %w", err)
	}
	err = runPatch(patch)
	if err != nil {
		return nil, fmt.Errorf("error applying patch: %w", err)
	}
	return patchBytes, nil
}

func RevertPatch(patch []byte) error {
	return common.RunCommandWithInput("patch -u --reverse -d app/", bytes.NewReader(patch))
}

func runPatch(patch []byte) error {
	return common.RunCommandWithInput("patch -ruN -F 10 -d app/", bytes.NewReader(patch))
}

func writePatchFile(patch []byte, fpath string) ([]byte, error) {
	newPatch := string(patch)
	if len(newPatch) == 0 || newPatch[len(newPatch)-1] != '\n' {
		newPatch += "\n"
	}

	// remove occurrences of ```diff and ``` from the patch
	newPatch = strings.ReplaceAll(newPatch, "```diff\n", "")
	newPatch = strings.ReplaceAll(newPatch, "```\n", "")
	newPatch = strings.ReplaceAll(newPatch, "<patch>\n", "")
	newPatch = strings.ReplaceAll(newPatch, "</patch>\n", "")

	err := os.WriteFile(fpath, []byte(newPatch), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing to patch file: %w", err)
	}

	return []byte(newPatch), nil
}
