package srp

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func Get_CongregationList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	coll := db.Collection("congregations")
	if coll == nil {
		log.Printf("Error: Database collection 'congregations' not found")
		http.Error(w, "Database collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	cur, err := coll.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("Error fetching congregations: %v", err)
		http.Error(w, "Error fetching congregations", http.StatusInternalServerError)
		return
	}

	var congregations []db.Congregation
	if err = cur.All(r.Context(), &congregations); err != nil {
		log.Printf("Error decoding congregations: %v", err)
		http.Error(w, "Error decoding congregations", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(congregations)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
