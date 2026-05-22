package srp

import (
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	srpID := r.PathValue("srp_id")

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	id, err := primitive.ObjectIDFromHex(srpID)
	if err != nil {
		log.Printf("Invalid SRP ID: %v", err)
		http.Error(w, "Invalid SRP ID", http.StatusBadRequest)
		return
	}

	res, err := collSRP.DeleteOne(r.Context(), bson.M{"_id": id})
	if err != nil {
		log.Println("Error deleting SRP:", err)
		http.Error(w, "Error deleting SRP", http.StatusInternalServerError)
		return
	}

	if res.DeletedCount == 0 {
		http.Error(w, "SRP not found", http.StatusNotFound)
		return
	}

	if res.DeletedCount > 1 {
		log.Printf("Warning: deleted %d documents for SRP ID %s", res.DeletedCount, srpID)
	}
}
