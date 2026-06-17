package users

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Post_Create tworzy nowe konto admin.
func Post_Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := guard(w, r); !ok {
		return
	}

	var req userReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// odrzuć duplikat loginu
	if _, err := db.GetUserByUsername(r.Context(), req.Login); err == nil {
		http.Error(w, "User with this login already exists", http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("users.Post_Create: hash: %v", err)
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	u := db.User{
		Username:    req.Login,
		Hash:        string(hash),
		FirstName:   req.Fn,
		LastName:    req.Ln,
		Permissions: req.Permissions,
	}
	id, err := db.CreateUser(r.Context(), &u)
	if err != nil {
		log.Printf("users.Post_Create: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id.Hex()})
}
