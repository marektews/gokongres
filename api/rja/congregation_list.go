package rja

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_CongregationList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type Request struct {
		TuraID int `json:"tura"`
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	coll := db.Collection("congregations")
	if coll == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	var congregations []db.Congregation
	sortOrder := bson.D{{Key: "lang", Value: 1}, {Key: "name", Value: 1}}
	opts := options.Find().SetSort(sortOrder)
	cursor, err := coll.Find(r.Context(), bson.M{"tura": req.TuraID}, opts)
	if err != nil {
		log.Printf("Error finding congregations for tura ID: %v, error: %v", req.TuraID, err)
		http.Error(w, "Error finding congregations", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	err = cursor.All(r.Context(), &congregations)
	if err != nil {
		log.Println("Error decoding congregations:", err)
		http.Error(w, "Error decoding congregations", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(congregations)
	if err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
