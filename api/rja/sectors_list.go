package rja

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_SectorsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("rja.GetSectorsList called")

	terminal_id := r.PathValue("terminal_id")
	log.Printf("Terminal ID: '%v'", terminal_id)

	oid, err := primitive.ObjectIDFromHex(terminal_id)
	if err != nil {
		log.Printf("Invalid terminal ID: %v", err)
		http.Error(w, "Invalid terminal ID", http.StatusBadRequest)
		return
	}

	coll := db.Collection("terminals")
	if coll == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	var terminal db.Terminal
	err = coll.FindOne(r.Context(), bson.M{"_id": oid}).Decode(&terminal)
	if err != nil {
		log.Printf("Error finding terminals: %v", err)
		http.Error(w, "Error finding terminals", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(terminal.Sectors)
	if err != nil {
		log.Printf("Error encoding sectors response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
