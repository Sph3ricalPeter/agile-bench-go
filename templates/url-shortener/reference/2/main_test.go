package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	TestUrl  = "https://www.google.com"
	TestSlug = "7378mDnD7gNvAj6WNbw7BW1u4ts="
)

func TestStoragePut(t *testing.T) {
	s := newTestStorage()
	slug := s.Put(TestUrl)
	if slug == "" {
		t.Error("expected non-empty slug")
	}
}

func TestStorageGet(t *testing.T) {
	s := newTestStorage()
	url, err := s.Get(TestSlug)
	if err != nil {
		t.Fatalf("failed to get slug: %v", err)
	}
	if url != TestUrl {
		t.Errorf("expected %v, got %v", TestUrl, url)
	}
}

func TestGetAlive(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/alive", nil)
	w := httptest.NewRecorder()
	h := newTestHandler()
	h.HandleGetAlive(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %v, got %v", http.StatusNoContent, resp.Status)
	}
}

func newTestStorage() *InMemUrlStorage {
	return &InMemUrlStorage{slugToUrlMap: map[string]string{
		TestSlug: TestUrl,
	}}
}

func newTestHandler() *ShortenHandler {
	return NewShortenHandler(newTestStorage())
}
