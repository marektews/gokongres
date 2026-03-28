package api

import (
	"net/http"

	"gokongres/api/arrivals"
	"gokongres/api/auth"
	"gokongres/api/limits"
	"gokongres/api/pk"
	"gokongres/sessions"

	"github.com/gorilla/mux"
)

// RegisterHandlers rejestruje endpointy HTTP używane przez serwera.
func RegisterHandlers(host string, port int) {
	r := mux.NewRouter()

	// Dodaj session middleware
	r.Use(sessions.SessionMiddleware)

	r.HandleFunc("/", RootHandler(host, port)).Methods(http.MethodGet)

	// Auth endpoints
	r.HandleFunc("/api/auth/login", auth.LoginHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/admin", auth.AdminHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/logout", auth.LogoutHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/auth/permissions", auth.PermissionsHandler).Methods(http.MethodPost)

	r.HandleFunc("/api/arrivals/all", arrivals.GetAll).Methods(http.MethodGet)
	r.HandleFunc("/api/arrivals/set", arrivals.Set).Methods(http.MethodPost)

	// limits endpoints
	r.HandleFunc("/api/limits/zbory", limits.GetZbory).Methods(http.MethodGet)
	r.HandleFunc("/api/limits/zbory/setlimit", limits.SetZboryLimit).Methods(http.MethodPost)
	r.HandleFunc("/api/limits/dzialy", limits.GetDzialy).Methods(http.MethodGet)
	r.HandleFunc("/api/limits/dzialy/setlimit", limits.SetDzialyLimit).Methods(http.MethodPost)

	// PK (parking księżycowy/torwar) endpoints
	r.HandleFunc("/api/pk/hints", pk.GetHints).Methods(http.MethodGet)
	r.HandleFunc("/api/pk/all", pk.GetLoadAll).Methods(http.MethodGet)

	// podłączenie routera do serwera HTTP
	http.Handle("/", r)
}
