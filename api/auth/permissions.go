package auth

import (
	"context"
	"encoding/json"
	"gokongres/db"
	"gokongres/sessions"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
* Pobiera uprawnienia aktualnie zalogowanego użytkownika.
* Endpoint ten może być używany przez frontend do sprawdzenia, jakie uprawnienia ma aktualnie zalogowany użytkownik.
* Odpowiedź zawiera zawartość struktury UserPermissions odpowiadającą uid użytkownika.
 */
func PermissionsHandler(w http.ResponseWriter, r *http.Request) {
	sessionData := sessions.GetSessionData(r)
	if sessionData == nil {
		log.Printf("PermissionsHandler: No active session")
		http.Error(w, "no active session", http.StatusUnauthorized)
		return
	}

	oid, err := primitive.ObjectIDFromHex(sessionData.UserID)
	if err != nil {
		log.Printf("PermissionsHandler: Invalid user ID format: %v", err)
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByID(context.Background(), oid)
	if err != nil {
		log.Printf("PermissionsHandler: Error fetching user for UID %s: %v", sessionData.UserID, err)
		http.Error(w, "error fetching user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user.Permissions); err != nil {
		log.Printf("PermissionsHandler: Error encoding permissions response: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
