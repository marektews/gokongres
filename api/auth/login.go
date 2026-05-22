package auth

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
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
		log.Printf("Error decoding login request: %v", err)
		http.Error(w, "invalid request body, JSON expected", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Passwd == "" {
		log.Println("Login or password not provided")
		http.Error(w, "login and passwd fields are required", http.StatusBadRequest)
		return
	}

	iPasswd, err := strconv.Atoi(creds.Passwd)
	if err != nil {
		log.Printf("Error converting password to integer: %v", err)
		http.Error(w, "invalid password format", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt for login '%s' (%d)", creds.Login, iPasswd)

	// wyszukiwanie zboru o podanej nazwie i numerze
	coll := db.Collection("congregations")
	if coll == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	res := coll.FindOne(r.Context(), bson.M{"name": creds.Login, "number": iPasswd})
	if res.Err() != nil {
		log.Printf("Login failed for login '%s': %v", creds.Login, res.Err())
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}

	var congregation db.Congregation
	err = res.Decode(&congregation)
	if err != nil {
		log.Printf("Error decoding congregation for login '%s': %v", creds.Login, err)
		http.Error(w, "Error decoding congregation for login", http.StatusUnauthorized)
		return
	}
	// logowanie udane nic więcej nie potrzeba oprócz statusu 200 OK
	log.Printf("Login successful for congregation '%s'", congregation.Name)
	w.WriteHeader(http.StatusOK)
}
