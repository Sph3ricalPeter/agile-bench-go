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

func splitIterative(s string, d string) []string {
	var parts []string
	var part string
	for _, c := range s {
		if strings.Contains(d, string(c)) {
			parts = append(parts, part)
			part = ""
		} else {
			part += string(c)
		}
	}
	parts = append(parts, part)
	return parts
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func binarySearch(arr []int, target int) int {
	low, high := 0, len(arr)-1
	for low <= high {
		mid := low + (high-low)/2
		if arr[mid] == target {
			return mid
		}
		if arr[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}

func merge(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	for len(left) > 0 || len(right) > 0 {
		if len(left) == 0 {
			return append(result, right...)
		}
		if len(right) == 0 {
			return append(result, left...)
		}
		if left[0] <= right[0] {
			result = append(result, left[0])
			left = left[1:]
		} else {
			result = append(result, right[0])
			right = right[1:]
		}
	}
	return result
}

func mergeSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}
	mid := len(arr) / 2
	left := mergeSort(arr[:mid])
	right := mergeSort(arr[mid:])
	return merge(left, right)
}

func minDistance(distances map[string]int, visited map[string]bool) string {
	min := 1<<31 - 1
	var node string
	for k, v := range distances {
		if !visited[k] && v < min {
			min = v
			node = k
		}
	}
	return node
}

func dijkstra(graph map[string]map[string]int, start string) map[string]int {
	distances := make(map[string]int)
	for node := range graph {
		distances[node] = 1<<31 - 1
	}
	distances[start] = 0
	visited := make(map[string]bool)
	for len(visited) < len(graph) {
		node := minDistance(distances, visited)
		visited[node] = true
		for neighbor, weight := range graph[node] {
			if distances[node]+weight < distances[neighbor] {
				distances[neighbor] = distances[node] + weight
			}
		}
	}
	return distances
}

func fibConc(n int) int {
	if n <= 1 {
		return n
	}
	ch := make(chan int, 2)
	go func() {
		ch <- fibConc(n - 1)
	}()
	go func() {
		ch <- fibConc(n - 2)
	}()
	return <-ch + <-ch
}
