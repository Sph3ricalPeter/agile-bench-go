package main

import (
	"bytes"
	"encoding/json"
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

func TestShortenTableDriven(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expStatus int
		expSlug   string
	}{
		{"valid url", TestUrl, http.StatusOK, TestSlug},
		{"invalid url", "invalid", http.StatusBadRequest, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/shorten",
				mustJsonEncode(t, ShortenRequest{Url: tt.url}))
			w := httptest.NewRecorder()
			h := newTestHandler()
			h.HandlePostShorten(w, r)
			resp := w.Result()
			if resp.StatusCode != tt.expStatus {
				t.Fatalf("expected status %v, got %v", tt.expStatus, resp.Status)
			}

			if tt.expStatus == http.StatusOK {
				var sr ShortenResponse
				mustDecodeResponseHelper(t, resp, &sr)
				if sr.ShortUrl != "" && sr.ShortUrl != fullUrl(tt.expSlug) {
					t.Errorf("expected %v, got %v", fullUrl(tt.expSlug), sr.ShortUrl)
				}
			}
		})
	}
}

func TestLookupTableDriven(t *testing.T) {
	tests := []struct {
		name      string
		slug      string
		expStatus int
	}{
		{"valid slug", TestSlug, http.StatusFound},
		{"invalid slug", "invalid", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tt.slug, nil)
			w := httptest.NewRecorder()
			h := newTestHandler()
			h.HandleLookup(w, r)
			resp := w.Result()
			if resp.StatusCode != tt.expStatus {
				t.Fatalf("expected status %v, got %v", tt.expStatus, resp.Status)
			}
		})
	}
}

func newTestHandler() *ShortenHandler {
	return NewShortenHandler(newTestStorage())
}

func newTestStorage() UrlStorage {
	s := NewInMemUrlStorage()
	s.Put(TestUrl)
	return s
}

func mustJsonEncode(t *testing.T, req interface{}) *bytes.Buffer {
	t.Helper()
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(req); err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}
	return body
}

func mustDecodeResponseHelper(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}
