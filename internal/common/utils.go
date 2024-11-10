package common

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

// MustGetEnv panics if the environment variable is not set. Env vars are loaded from .env file.
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Errorf("%s environment variable is required", key))
	}
	return value
}

// Must panics if err is not nil, otherwise returns the value.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(fmt.Errorf("unexpected error: %w", err))
	}
	return val
}

// CheckErr panics if err is not nil.
func CheckErr(err error) {
	if err != nil {
		panic(fmt.Errorf("unexpected error: %w", err))
	}
}
