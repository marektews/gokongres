package rja

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_BusesUsed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tura_id := r.PathValue("tura_id")

	// zbory przypisane do tej tury (zbiór ID dla szybkiego sprawdzania przynależności)
	congregations, err := db.GetCongregationsForTura(r.Context(), tura_id)
	if err != nil {
		log.Printf("Get_BusesUsed: error getting congregations for tura %s: %v", tura_id, err)
		http.Error(w, "Error getting congregations for tura", http.StatusInternalServerError)
		return
	}
	turaCongs := make(map[primitive.ObjectID]bool, len(congregations))
	for _, c := range congregations {
		turaCongs[c.ID] = true
	}

	collRJA := db.Collection("rja")
	collSRA := db.Collection("sra")
	if collRJA == nil || collSRA == nil {
		log.Println("Get_BusesUsed: required collection ('rja' or 'sra') not found")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	cur, err := collRJA.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("Get_BusesUsed: error finding buses for tura %s: %v", tura_id, err)
		http.Error(w, "Error finding buses for tura", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var buses []db.RJA
	if err = cur.All(r.Context(), &buses); err != nil {
		log.Printf("Get_BusesUsed: error decoding buses for tura %s: %v", tura_id, err)
		http.Error(w, "Error decoding buses for tura", http.StatusInternalServerError)
		return
	}

	// płaska lista identyfikatorów SRA użytych w rozkładzie jazdy danej tury
	// (zgodnie ze starym API: frontend sprawdza przynależność przez Array.includes)
	resp := make([]string, 0)
	for _, bus := range buses {
		var sra db.SRA
		if err := collSRA.FindOne(r.Context(), bson.M{"_id": bus.SraID}).Decode(&sra); err != nil {
			continue // brak powiązanego SRA → pomijamy
		}
		if turaCongs[sra.CongregationID] {
			resp = append(resp, bus.SraID.Hex())
		}
	}

	if err = json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Get_BusesUsed: error encoding response for tura %s: %v", tura_id, err)
		http.Error(w, "Error encoding response for tura", http.StatusInternalServerError)
		return
	}
}
