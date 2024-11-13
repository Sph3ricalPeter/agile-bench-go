package main

import (
	"net/http"
)

func handleGetAlive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	http.HandleFunc("/api/v1/alive", handleGetAlive)
	http.ListenAndServe(":8080", nil)
}
