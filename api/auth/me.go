package auth

import (
	"encoding/json"
	"net/http"

	"gokongres/sessions"
)

// MeHandler zwraca informacje o zalogowanym użytkowniku
func MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionData := sessions.GetSessionData(r)
	if sessionData == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Not logged in",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"user":    sessionData,
	})
}
