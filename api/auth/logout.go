package auth

import (
	"encoding/json"
	"gokongres/sessions"
	"net/http"
)

// LogoutHandler obsługuje wylogowanie użytkownika
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := sessions.ClearSession(w, r); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"success": false,
			"message": "Failed to logout",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Logged out successfully",
	})
}
