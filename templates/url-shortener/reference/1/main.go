package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

const (
	Host = "localhost"
	Port = "8080"
)

type UrlStorage interface {
	// takes a URL to shorten and returns a slug, e.g. https://google.com/ -> abc123
	Put(url string) string

	// takes a slug and returns the original URL, e.g. abc123 -> https://google.com/
	Get(slug string) (string, error)
}

type InMemUrlStorage struct {
	slugToUrlMap map[string]string
}

func NewInMemUrlStorage() *InMemUrlStorage {
	return &InMemUrlStorage{
		slugToUrlMap: make(map[string]string),
	}
}

func (s *InMemUrlStorage) Put(url string) string {
	slug := hash(url)
	s.slugToUrlMap[slug] = url
	return slug
}

func (s *InMemUrlStorage) Get(slug string) (string, error) {
	if url, ok := s.slugToUrlMap[slug]; ok {
		return url, nil
	}
	return "", errors.New("not found")
}

type ShortenHandler struct {
	Storage UrlStorage
}

func NewShortenHandler(storage UrlStorage) *ShortenHandler {
	return &ShortenHandler{Storage: storage}
}

func (h *ShortenHandler) HandleGetAlive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *ShortenHandler) HandlePostShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := validateURL(req.Url); err != nil {
		log.Printf("invalid url: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slug := h.Storage.Put(req.Url)
	fullUrl := fullUrl(slug)

	log.Printf("shortened %s to %s", req.Url, fullUrl)
	resp := ShortenResponse{ShortUrl: fullUrl}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ShortenHandler) HandleLookup(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[1:]
	log.Printf("looking up slug %s", slug)
	url, err := h.Storage.Get(slug)
	if err != nil {
		log.Printf("failed to get url: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

type ShortenRequest struct {
	Url string `json:"url"`
}

type ShortenResponse struct {
	ShortUrl string `json:"short_url"`
}

func main() {
	s := NewInMemUrlStorage()
	h := NewShortenHandler(s)
	http.HandleFunc("GET /alive", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	http.HandleFunc("POST /shorten", h.HandlePostShorten)
	http.HandleFunc("GET /", h.HandleLookup)
	http.ListenAndServe(fmt.Sprintf("%s:%s", Host, Port), nil)
}

func fullUrl(slug string) string {
	return fmt.Sprintf("http://%s:%s/%s", Host, Port, slug)
}

func validateURL(url string) error {
	if url == "" {
		return errors.New("url is empty")
	}
	re := regexp.MustCompile(`^https?://`)
	if !re.MatchString(url) {
		return errors.New("url is not valid")
	}
	return nil
}

func hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	slug := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return slug
}
