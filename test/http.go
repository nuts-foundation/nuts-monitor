package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BasicTestNode(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "UP"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(func() {
		ts.Close()
	})
	return ts
}
