package api

import (
	"net/http"

	"gokongres/api/arrivals"

	"github.com/gorilla/mux"
)

// RegisterHandlers rejestruje endpointy HTTP używane przez serwera.
func RegisterHandlers(host string, port int) {
	r := mux.NewRouter()

	r.HandleFunc("/", RootHandler(host, port)).Methods(http.MethodGet)

	r.HandleFunc("/api/arrivals/all", arrivals.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/api/arrivals/set", arrivals.Set).Methods(http.MethodPost)

	http.Handle("/", r)
}
