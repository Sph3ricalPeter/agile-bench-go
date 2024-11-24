package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// MustWriteJsonFile writes data to a json file and creates all directories in the file path if they don't exist.
// Panics if there is an error writing the file.
func MustWriteJsonFile(data any, fpath string) {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Errorf("error marshalling data: %w", err))
	}

	err = os.MkdirAll(path.Dir(fpath), os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("error creating directories: %w", err))
	}

	err = os.WriteFile(fpath, bytes, 0644)
	if err != nil {
		panic(fmt.Errorf("error writing data: %w", err))
	}
}

// MustReadJsonFile reads a json file and returns the unmarshaled data.
func MustReadJsonFile(fpath string) any {
	bytes, err := os.ReadFile(fpath)
	if err != nil {
		panic(fmt.Errorf("error reading file: %w", err))
	}

	var data any
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling data: %w", err))
	}

	return data
}

// MustReadJsonFileInto reads a json file and unmarshals the data into the given type using generics.
func MustReadJsonFileInto[T any](fpath string) T {
	bytes, err := os.ReadFile(fpath)
	if err != nil {
		panic(fmt.Errorf("error reading file: %w", err))
	}

	var data T
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(fmt.Errorf("error unmarshalling data: %w", err))
	}

	return data
}
