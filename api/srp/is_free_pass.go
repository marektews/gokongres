package srp

import (
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

/**
* Sprawdzanie czy jest jeszcze wolny identyfikator do wykorzystania
* return: HTTP codes: 200 | 404
 */
func Get_IsFreePass(w http.ResponseWriter, r *http.Request) {
	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	congregationName := r.PathValue("congregation_name")

	var congregation db.Congregation
	err := collCongr.FindOne(r.Context(), bson.M{"name": congregationName}).Decode(&congregation)
	if err != nil {
		log.Println("Error occurred while finding congregation:", err)
		http.Error(w, "Error occurred while finding congregation", http.StatusInternalServerError)
		return
	}

	count, err := collSRP.CountDocuments(r.Context(), bson.M{"congregation_id": congregation.ID})
	if err != nil {
		log.Println("Error occurred while counting SRP documents:", err)
		http.Error(w, "Error occurred while counting SRP documents", http.StatusInternalServerError)
		return
	}

	if count < int64(congregation.Plimit) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
