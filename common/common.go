package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Errorf("%s environment variable is required", key))
	}
	return value
}

func WritePatchFile(patch []byte, fpath string) ([]byte, error) {
	newPatch := string(patch)
	if len(newPatch) == 0 || newPatch[len(newPatch)-1] != '\n' {
		newPatch += "\n"
	}

	// remove occurrences of ```diff and ``` from the patch
	newPatch = strings.ReplaceAll(newPatch, "```diff\n", "")
	newPatch = strings.ReplaceAll(newPatch, "```\n", "")

	err := os.WriteFile(fpath, []byte(newPatch), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing to patch file: %w", err)
	}

	return []byte(newPatch), nil
}

// CreateCacheKey returns the SHA256 hash of the given data prefixed with a number like "1_sgeas234..."
func CreateCacheKey(data []byte, number int) string {
	s := sha256.New()
	s.Write(data)
	return fmt.Sprintf("%d_%s", number, hex.EncodeToString(s.Sum(nil)))
}

// Must panics if err is not nil, otherwise returns the value.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// CheckErr panics if err is not nil.
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
