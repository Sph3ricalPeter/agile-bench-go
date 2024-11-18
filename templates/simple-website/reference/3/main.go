package main

import (
	"net/http"
	"text/template"
)

func handleGetAlive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func handleGetIndex(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/api/v1/alive", handleGetAlive)
	http.HandleFunc("/", handleGetIndex)
	http.ListenAndServe(":8080", nil)
}
