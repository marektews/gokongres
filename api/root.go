package api

import (
	"log"
	"net/http"
)

// RootHandler zwraca http.HandlerFunc obsługujący pattern '/' i inne brakujące ścieżki.
func RootHandler(host string, port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Unknown URL: %s %s from %s:%d", r.Method, r.URL.Path, host, port)
		http.NotFound(w, r)
	}
}
