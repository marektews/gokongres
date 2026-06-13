package pk

import (
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_DeletePass(w http.ResponseWriter, r *http.Request) {
	pk_id := r.PathValue("pk_id")
	pkID, err := primitive.ObjectIDFromHex(pk_id)
	if err != nil {
		log.Printf("Invalid pk_id format: %v", err)
		http.Error(w, "Invalid pk_id format", http.StatusBadRequest)
		return
	}

	coll := db.Collection("departments_pk")
	if coll == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	res, err := coll.DeleteOne(r.Context(), bson.M{"_id": pkID})
	if err != nil {
		log.Printf("Error deleting PK '%s': %v", pk_id, err)
		http.Error(w, "Error deleting PK", http.StatusInternalServerError)
		return
	}
	if res.DeletedCount == 0 {
		http.Error(w, "PK not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
