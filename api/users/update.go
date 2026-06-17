package users

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Post_Update aktualizuje konto admin. Hasło zmieniane tylko gdy podane (niepuste).
func Post_Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := guard(w, r); !ok {
		return
	}

	var req userReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	id, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	if req.Login == "" {
		http.Error(w, "Login is required", http.StatusBadRequest)
		return
	}

	set := bson.M{
		"username":    req.Login,
		"first_name":  req.Fn,
		"last_name":   req.Ln,
		"permissions": req.Permissions,
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("users.Post_Update: hash: %v", err)
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		set["hash"] = string(hash)
	}

	if err := db.UpdateUser(r.Context(), id, set); err != nil {
		log.Printf("users.Post_Update: %v", err)
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
