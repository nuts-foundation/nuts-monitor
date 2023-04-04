package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestNode struct {
	server *httptest.Server
	mux    *http.ServeMux
}

func BasicTestNode(t *testing.T) TestNode {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP"}`))
	})

	ts := httptest.NewServer(mux)

	t.Cleanup(func() {
		ts.Close()
	})
	return TestNode{
		server: ts,
		mux:    mux,
	}
}

func (tn TestNode) HandleFunc(pattern string, handler http.HandlerFunc) {
	tn.mux.HandleFunc(pattern, handler)
}

func (tn TestNode) URL() string {
	return tn.server.URL
}
