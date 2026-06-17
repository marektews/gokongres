package users

import (
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Get_Delete usuwa konto admin (poza własnym).
func Get_Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	me, ok := guard(w, r)
	if !ok {
		return
	}

	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	if id == me.ID {
		http.Error(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	deleted, err := db.DeleteUser(r.Context(), id)
	if err != nil {
		log.Printf("users.Get_Delete: %v", err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}
	if deleted == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
