package sector

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func States(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sectorID := r.PathValue("sector_id")

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

	soaSortOrder := bson.D{{Key: "ts", Value: -1}, {Key: "_id", Value: -1}}
	soaOpts := options.FindOne().SetSort(soaSortOrder).SetCollation(&collation)

	rjaSortOrder := bson.D{{Key: "sector_id", Value: 1}, {Key: "sector_order", Value: 1}}
	rjaOpts := options.Find().SetSort(rjaSortOrder).SetCollation(&collation)
	cur, err := rjaColl.Find(r.Context(), bson.M{"sector_id": sectorID}, rjaOpts)
	if err != nil {
		log.Printf("Error finding RJAs for sector_id '%s': %v", sectorID, err)
		http.Error(w, "Error finding RJAs", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	type States struct {
		Status string `json:"status"`
		Ts     string `json:"ts"`
	}
	type SectorInfo struct {
		SectorID string   `json:"sector_id"`
		States   []States `json:"states"`
	}

	res := SectorInfo{SectorID: sectorID, States: []States{}}
	for cur.Next(r.Context()) {
		var rja db.RJA
		if err := cur.Decode(&rja); err != nil {
			log.Printf("Error decoding RJA for sector_id '%s': %v", sectorID, err)
			http.Error(w, "Error decoding RJA", http.StatusInternalServerError)
			return
		}

		if rja.WasArrived() {
			var soa db.SOA
			err = soaColl.FindOne(r.Context(), bson.M{"rja_id": rja.ID}, soaOpts).Decode(&soa)
			if err != nil {
				log.Printf("Error finding SOA for rja_id '%s': %v", rja.ID, err)
				http.Error(w, "Error finding SOA", http.StatusInternalServerError)
				continue
			}

			state := States{Status: soa.Status, Ts: soa.Timestamp.Format("02.01.2006 15:04:05")}
			res.States = append(res.States, state)
		}
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Printf("Error encoding response for sector_id '%s': %v", sectorID, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
