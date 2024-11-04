package common

import (
	"fmt"
	"os"
)

func ReqGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Printf("Error: %s environment variable is required\n", key)
		os.Exit(1)
	}
	return value
}
