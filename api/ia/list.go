package ia

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	congregationName := r.PathValue("congregation_name")

	collCong := db.Collection("congregations")
	if collCong == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	var cong db.Congregation
	if err := collCong.FindOne(r.Context(), bson.M{"name": congregationName}).Decode(&cong); err != nil {
		log.Printf("IA list: congregation '%s' not found: %v", congregationName, err)
		http.Error(w, "Congregation not found", http.StatusNotFound)
		return
	}

	collSRA := db.Collection("sra")
	if collSRA == nil {
		log.Println("Collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}

	cur, err := collSRA.Find(r.Context(), bson.M{
		"congregation_id": cong.ID,
		"bus":             bson.M{"$exists": true},
		"canceled":        bson.M{"$ne": true},
	})
	if err != nil {
		log.Printf("IA list: error finding SRA for congregation '%s': %v", congregationName, err)
		http.Error(w, "Error finding SRA", http.StatusInternalServerError)
		return
	}

	var sras []db.SRA
	if err := cur.All(r.Context(), &sras); err != nil {
		log.Printf("IA list: error decoding SRA: %v", err)
		http.Error(w, "Error decoding SRA", http.StatusInternalServerError)
		return
	}

	collPilots := db.Collection("pilots")
	if collPilots == nil {
		log.Println("Collection 'pilots' not found")
		http.Error(w, "Collection 'pilots' not found", http.StatusInternalServerError)
		return
	}

	getPilot := func(id primitive.ObjectID) *pilotInfo {
		var p db.Pilot
		if err := collPilots.FindOne(r.Context(), bson.M{"_id": id}).Decode(&p); err != nil {
			log.Printf("IA list: pilot %s not found: %v", id.Hex(), err)
			return nil
		}
		return &pilotInfo{Fn: p.FirstName, Ln: p.LastName}
	}

	result := make([]item, 0)
	for _, sra := range sras {
		it := item{
			ID:  sra.ID.Hex(),
			Bus: busInfo{Type: sra.Bus.Type},
		}
		if p := getPilot(sra.Pilot1ID); p != nil {
			it.Pilot1 = *p
		}
		if sra.Pilot2ID != nil {
			it.Pilot2 = getPilot(*sra.Pilot2ID)
		}
		if sra.Pilot3ID != nil {
			it.Pilot3 = getPilot(*sra.Pilot3ID)
		}
		result = append(result, it)
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("IA list: error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

type pilotInfo struct {
	Fn string `json:"fn"`
	Ln string `json:"ln"`
}
type busInfo struct {
	Type string `json:"type"`
}
type item struct {
	ID     string     `json:"id"`
	Bus    busInfo    `json:"bus"`
	Pilot1 pilotInfo  `json:"pilot1"`
	Pilot2 *pilotInfo `json:"pilot2,omitempty"`
	Pilot3 *pilotInfo `json:"pilot3,omitempty"`
}
