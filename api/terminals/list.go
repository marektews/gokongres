package terminals

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_AllList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("terminals.GetAllList called")

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

	// lista terminali do zwrócenia (bez sektorów)
	type TerminalResponse struct {
		ID   primitive.ObjectID `json:"tid"`
		Name string             `json:"name"`
	}
	terminals := []TerminalResponse{}
	for cur.Next(r.Context()) {
		var t db.Terminal
		if err := cur.Decode(&t); err != nil {
			log.Printf("Error decoding terminal: %v", err)
			http.Error(w, "Error decoding terminal", http.StatusInternalServerError)
			return
		}
		terminals = append(terminals, TerminalResponse{ID: t.ID, Name: t.Name})
	}

	if err := json.NewEncoder(w).Encode(terminals); err != nil {
		log.Printf("Error encoding terminals response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
