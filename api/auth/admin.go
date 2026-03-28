package auth

import (
	"context"
	"encoding/json"
	"gokongres/db"
	"gokongres/sessions"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

/**
* Logowanie na poziomie moderatora.
 */
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Printf("AdminHandler called with method %s", r.Method)

	// parsowanie zapytania
	type Credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Printf("AdminHandler: Error decoding request body: %v", err)
		http.Error(w, "invalid request body, JSON expected", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Password == "" {
		log.Printf("AdminHandler: Missing login or password in request")
		http.Error(w, "login and passwd fields are required", http.StatusBadRequest)
		return
	}

	// Pobierz użytkownika z bazy danych
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := db.GetUserByUsername(ctx, creds.Login)
	if err != nil {
		log.Printf("AdminHandler: User not found - %s", creds.Login)
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Weryfikuj hasło
	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(creds.Password)); err != nil {
		log.Printf("AdminHandler: Invalid credentials for user - %s", creds.Login)
		http.Error(w, "invalid credentials", http.StatusForbidden)
		return
	}

	log.Printf("AdminHandler: User found - %s", user.Username)

	// Utwórz dane sesji z danych pobranych z bazy
	sessionData := sessions.SessionData{
		UserID:   user.Uid,
		Username: user.Username,
	}

	// Zapisz sesję
	if err := sessions.SetSessionData(w, r, sessionData); err != nil {
		log.Printf("AdminHandler: Failed to create session for user - %s: %v", creds.Login, err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Przygotuj odpowiedź z danymi użytkownika
	type Response struct {
		Id          int                `json:"id"`
		FirstName   string             `json:"fn"`
		LastName    string             `json:"ln"`
		Permissions db.UserPermissions `json:"permissions"`
	}
	resp := Response{
		Id:          user.Uid,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Permissions: user.Permissions,
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("AdminHandler: Error encoding response: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
