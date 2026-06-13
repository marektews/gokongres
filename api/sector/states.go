package sector

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func States(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sector_id := r.PathValue("sector_id")
	oid, err := primitive.ObjectIDFromHex(sector_id)
	if err != nil {
		log.Printf("Invalid sector ID '%s': %v", sector_id, err)
		http.Error(w, "Invalid sector ID", http.StatusBadRequest)
		return
	}

	rjaColl := db.Collection("rja")
	if rjaColl == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}

	soaColl := db.Collection("soa")
	if soaColl == nil {
		log.Println("Collection 'soa' not found")
		http.Error(w, "Collection 'soa' not found", http.StatusInternalServerError)
		return
	}

	collation := options.Collation{Locale: "pl", NumericOrdering: true, Strength: 1}
	rjaOpts := options.Find().SetSort(bson.D{{Key: "sector_order", Value: 1}}).SetCollation(&collation)
	cur, err := rjaColl.Find(r.Context(), bson.M{"sector_id": oid}, rjaOpts)
	if err != nil {
		log.Printf("Error finding RJAs for sector_id '%s': %v", sector_id, err)
		http.Error(w, "Error finding RJAs", http.StatusInternalServerError)
		return
	}

	var allRJA []db.RJA
	if err := cur.All(r.Context(), &allRJA); err != nil {
		log.Printf("Error decoding RJAs for sector_id '%s': %v", sector_id, err)
		http.Error(w, "Error decoding RJAs", http.StatusInternalServerError)
		return
	}

	type State struct {
		Status string `json:"status"`
		Ts     string `json:"ts"`
	}
	type Response struct {
		SectorID primitive.ObjectID `json:"sid"`
		States   map[string]State   `json:"states"` // inicjalizacja {} → nie null
	}

	resp := Response{
		SectorID: oid,
		States:   map[string]State{},
	}

	for _, rja := range allRJA {
		if !rja.WasArrived() {
			continue
		}

		var soa db.SOA
		err = soaColl.FindOne(r.Context(), bson.M{"rja_id": rja.ID}).Decode(&soa)
		if err != nil {
			log.Printf("Error finding SOA for rja_id '%s': %v", rja.ID.Hex(), err)
			continue // brak dokumentu SOA → pomijamy ten autobus
		}

		last, ok := soa.Latest()
		if !ok {
			continue // dokument bez stanów → pomijamy
		}

		resp.States[rja.ID.Hex()] = State{
			Status: last.State,
			Ts:     last.Ts.Format("02.01.2006 15:04:05"),
		}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response for sector_id '%s': %v", sector_id, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
