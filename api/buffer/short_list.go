package buffer

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_AllShortList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("buffer.GetAllShortList called")

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

	var terminals []db.Terminal
	if err := cur.All(r.Context(), &terminals); err != nil {
		log.Printf("Error decoding terminals: %v", err)
		http.Error(w, "Error decoding terminals", http.StatusInternalServerError)
		return
	}

	type TerminalInfo struct {
		ID   primitive.ObjectID `json:"id"`
		Name string             `json:"name"`
	}
	var terminalInfos []TerminalInfo
	for _, t := range terminals {
		terminalInfos = append(terminalInfos, TerminalInfo{
			ID:   t.ID,
			Name: t.Name,
		})
	}

	if err := json.NewEncoder(w).Encode(terminalInfos); err != nil {
		log.Printf("Error encoding terminals info to JSON: %v", err)
		http.Error(w, "Error encoding terminals info to JSON", http.StatusInternalServerError)
		return
	}
}
