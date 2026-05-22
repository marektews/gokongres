package rja

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func Get_BusesSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("rja.GetBusesSave called")

	type RJAEntry struct {
		db.RJA
		Canceled bool `json:"canceled"`
	}

	type Request struct {
		SectorID string     `json:"sid"`
		TuraID   string     `json:"tura"`
		Buses    []RJAEntry `json:"rja"`
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding request for sector %s and tura %s: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	coll := db.Collection("rja")
	if coll == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}

	// kasowanie poprzedniej listy autobusów dla tej tury
	_, err = coll.DeleteMany(r.Context(), bson.M{"sector_id": req.SectorID, "tura_id": req.TuraID})
	if err != nil {
		log.Printf("Error deleting old buses for sector %s and tura %s: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error deleting old buses for sector and tura", http.StatusInternalServerError)
		return
	}

	// wstawienie nowej listy autobusów
	buses := make([]any, len(req.Buses))
	for i, bus := range req.Buses {
		buses[i] = bus
	}
	_, err = coll.InsertMany(r.Context(), buses)
	if err != nil {
		log.Printf("Error inserting new buses for sector %s and tura %s: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error inserting new buses for sector and tura", http.StatusInternalServerError)
		return
	}
}
