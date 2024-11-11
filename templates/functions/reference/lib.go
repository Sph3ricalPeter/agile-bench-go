package main

import "strings"

func add(a, b int) int {
	return a + b
}

func subtract(a, b int) int {
	return a - b
}

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func split(s string, d string) []string {
	return strings.Split(s, d)
}

func splitCustom(s string, d string) []string {
	var result []string
	for _, v := range strings.Split(s, d) {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}
