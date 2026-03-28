package auth

import (
	"context"
	"encoding/json"
	"gokongres/db"
	"gokongres/sessions"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

/**
* Logowanie za pomocą nazwy i numeru zboru.
* Obsługuje logowanie użytkownika.
 */
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// parsowanie zapytania
	type Credentials struct {
		Login  string `json:"login"`
		Passwd string `json:"passwd"` //TODO: zmienić na password (może)
	}

	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "invalid request body, JSON expected", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Passwd == "" {
		http.Error(w, "login and passwd fields are required", http.StatusBadRequest)
		return
	}

	// Pobierz użytkownika z bazy danych
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := db.GetUserByUsername(ctx, creds.Login)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Weryfikuj hasło
	if err := bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(creds.Passwd)); err != nil {
		http.Error(w, "invalid credentials", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Utwórz dane sesji z danych pobranych z bazy
	sessionData := sessions.SessionData{
		UserID:   user.ID,
		Username: user.Username,
	}

	// Zapisz sesję
	if err := sessions.SetSessionData(w, r, sessionData); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
}
