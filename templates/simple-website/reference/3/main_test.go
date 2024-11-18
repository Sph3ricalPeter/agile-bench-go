package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
)

func TestApi(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/alive", nil)
	rr := httptest.NewRecorder()
	h := http.HandlerFunc(handleGetAlive)
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h := http.HandlerFunc(handleGetIndex)
	h.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// parse html to find h1
	doc, err := htmlquery.Parse(rr.Body)
	if err != nil {
		t.Fatalf("failed to load html: %v", err)
	}
	got := htmlquery.FindOne(doc, "//h1/text()").Data
	want := "Welcome to simple website!"
	if got != want {
		t.Fatalf("want %s, got %s", got, want)
	}

	// test for bootstrap css import
	link := htmlquery.FindOne(doc, "//link[@rel=\"stylesheet\"]")
	want = "https://cdn.jsdelivr.net/npm/bootstrap"
	got = htmlquery.SelectAttr(link, "href")
	if !strings.Contains(got, want) {
		t.Fatalf("want contains %s, got %s", want, got)
	}
}
