package rja

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_TerminalsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("rja.GetTerminalsList called")

	coll := db.Collection("terminals")
	if coll == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	collation := options.Collation{Locale: "pl", NumericOrdering: true, Strength: 1}
	sortOrder := bson.D{{Key: "name", Value: 1}}
	opts := options.Find().SetSort(sortOrder).SetCollation(&collation)

	cur, err := coll.Find(r.Context(), bson.M{}, opts)
	if err != nil {
		log.Printf("Error finding terminals: %v", err)
		http.Error(w, "Error finding terminals", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	// przygotowywanie listy terminali do zwrócenia
	type TerminalResponse struct {
		ID   primitive.ObjectID `json:"tid"`
		Name string             `json:"name"`
	}
	var terminals []TerminalResponse
	for cur.Next(r.Context()) {
		var terminal db.Terminal
		if err := cur.Decode(&terminal); err != nil {
			log.Printf("Error decoding terminal: %v", err)
			http.Error(w, "Error decoding terminal", http.StatusInternalServerError)
			return
		}
		terminals = append(terminals, TerminalResponse{
			ID:   terminal.ID,
			Name: terminal.Name,
		})
	}

	err = json.NewEncoder(w).Encode(terminals)
	if err != nil {
		log.Printf("Error encoding terminals response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
