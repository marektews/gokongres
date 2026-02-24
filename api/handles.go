package api

import "net/http"

// RegisterHandlers rejestruje endpointy HTTP używane przez serwera.
func RegisterHandlers(host string, port int) {
    http.HandleFunc("/", RootHandler(host, port))
}
