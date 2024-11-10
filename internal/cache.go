package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/Sph3ricalPeter/frbench/internal/common"
)

type JsonCache struct {
	BasePath string
}

func NewJsonCache(basePath string) *JsonCache {
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

// CreateCacheKey returns the SHA256 hash of the given data prefixed with a number like "1_sgeas234..."
func CreateCacheKey(data []byte, number int) string {
	s := sha256.New()
	s.Write(data)
	return fmt.Sprintf("%d_%s", number, hex.EncodeToString(s.Sum(nil)))
}
