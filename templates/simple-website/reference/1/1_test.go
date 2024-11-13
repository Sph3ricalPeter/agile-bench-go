package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
