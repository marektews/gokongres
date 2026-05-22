package sra

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func Get_SearchCongregationsByPattern(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	pattern := r.PathValue("pattern")
	log.Printf("Searching by pattern: %s", pattern)

	coll := db.Collection("congregations")
	if coll == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}
	cur, err := coll.Find(r.Context(), bson.M{"name": bson.M{"$regex": pattern, "$options": "i"}})
	if err != nil {
		log.Printf("Error searching congregations with pattern '%s': %v", pattern, err)
		http.Error(w, "Error searching congregations", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var congregationNames []string
	for cur.Next(r.Context()) {
		var congregation db.Congregation
		err := cur.Decode(&congregation)
		if err != nil {
			log.Printf("Error decoding congregation: %v", err)
			http.Error(w, "Error decoding congregation", http.StatusInternalServerError)
			return
		}
		congregationNames = append(congregationNames, congregation.Name)
	}
	log.Printf("Searched congregations: %+v", congregationNames)

	err = json.NewEncoder(w).Encode(congregationNames)
	if err != nil {
		log.Printf("Error encoding response for congregations with pattern '%s': %v", pattern, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
