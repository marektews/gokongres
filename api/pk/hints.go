package pk

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_Hints(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	coll := db.Collection("departments")
	if coll == nil {
		log.Print("GetHints: mongo client not initialized")
		http.Error(w, "Mongo client not initialized", http.StatusInternalServerError)
		return
	}

	sortOrder := bson.D{{Key: "tura", Value: 1}, {Key: "lang", Value: 1}, {Key: "name", Value: 1}}
	opts := options.Find().SetSort(sortOrder)
	cursor, err := coll.Find(r.Context(), bson.M{}, opts)
	if err != nil {
		log.Printf("GetHints: Error fetching hints: %v", err)
		http.Error(w, "Error fetching hints: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var departments []db.Department
	if err := cursor.All(r.Context(), &departments); err != nil {
		log.Printf("GetHints: Error decoding departments: %v", err)
		http.Error(w, "Error decoding departments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type Hint struct {
		ID     primitive.ObjectID `json:"id"`
		Name   string             `json:"name"`
		Lang   string             `json:"lang"`
		Tura   int                `json:"tura"`
		Plimit int                `json:"plimit"`
	}
	hints := make([]Hint, 0)
	for _, dept := range departments {
		hints = append(hints, Hint{
			ID:     dept.ID,
			Name:   dept.Name,
			Lang:   dept.Lang,
			Tura:   dept.TuraID,
			Plimit: dept.Plimit,
		})
	}

	err = json.NewEncoder(w).Encode(hints)
	if err != nil {
		log.Printf("GetHints: Error encoding hints to JSON: %v", err)
		http.Error(w, "Error encoding hints to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
