package rja

import (
	"context"
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RJAEntry struct {
	db.RJA
	Canceled bool `json:"canceled"`
}

func Get_BusesSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("rja.GetBusesSave called")

	type Request struct {
		SectorID string     `json:"sid"`
		TuraID   int        `json:"tura_id"`
		Buses    []RJAEntry `json:"rja"`
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding request for sector %s and tura %d: %v", req.SectorID, req.TuraID, err)
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
	busIDs := getBusesIDs(req.Buses)
	// log.Println("Bus IDs to delete:", busIDs)
	_, err = coll.DeleteMany(r.Context(), bson.M{"_id": bson.M{"$in": busIDs}})
	if err != nil {
		log.Printf("Error deleting old buses for sector %s and tura %d: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error deleting old buses for sector and tura", http.StatusInternalServerError)
		return
	}

	// wstawienie nowej listy autobusów
	buses := make([]any, len(req.Buses))
	for i, bus := range req.Buses {
		buses[i] = bus.RJA
		updateSRACanceledStatus(r.Context(), bus.SraID, bus.Canceled)
	}
	_, err = coll.InsertMany(r.Context(), buses)
	if err != nil {
		log.Printf("Error inserting new buses for sector %s and tura %d: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error inserting new buses for sector and tura", http.StatusInternalServerError)
		return
	}
}

/**
 * Aktualizacja statusu anulowania SRA
 */
func updateSRACanceledStatus(ctx context.Context, sraID primitive.ObjectID, canceled bool) {
	coll := db.Collection("sra")
	if coll == nil {
		log.Println("Collection 'sra' not found")
		return
	}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": sraID}, bson.M{"$set": bson.M{"canceled": canceled}})
	if err != nil {
		log.Printf("Error updating SRA canceled status for ID %s: %v", sraID.Hex(), err)
	}
}

/**
 * Tworzenie listy ID autobusów z listy autobusów
 */
func getBusesIDs(buses []RJAEntry) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, len(buses))
	for i, bus := range buses {
		ids[i] = bus.ID
	}
	return ids
}
