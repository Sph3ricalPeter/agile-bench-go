package internal

import (
	"fmt"
	"os"

	"github.com/Sph3ricalPeter/frbench/common"
)

type JsonCache struct {
	BasePath string
}

func NewJsonCache(basePath string) *JsonCache {
	// create the cache directory if it doesn't exist
	common.CheckErr(os.MkdirAll(basePath, 0755))
	return &JsonCache{BasePath: basePath}
}

func (j *JsonCache) Get(key string) ([]byte, bool) {
	data, err := os.ReadFile(fmt.Sprintf("%s/%s.json", j.BasePath, key))
	if err != nil {
		return nil, false
	}
	return data, true
}

func (j *JsonCache) Put(key string, data []byte) error {
	err := os.WriteFile(fmt.Sprintf("%s/%s.json", j.BasePath, key), data, 0644)
	if err != nil {
		return fmt.Errorf("error writing cache file: %w", err)
	}
	return nil
}

func (j *JsonCache) Delete(key string) error {
	err := os.Remove(fmt.Sprintf("%s/%s.json", j.BasePath, key))
	if err != nil {
		return fmt.Errorf("error removing cache file: %w", err)
	}
	return nil
}

func (j *JsonCache) Clear() error {
	err := os.RemoveAll(j.BasePath)
	if err != nil {
		return fmt.Errorf("error clearing cache: %w", err)
	}
	return nil
}
