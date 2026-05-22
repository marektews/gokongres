package rja

import (
	"encoding/json"
	"gokongres/db"
	"gokongres/helpers"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_BusesOfSector(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("rja.GetBusesOfSector called")

	tura_id := r.PathValue("tura_id")
	turaID, err := strconv.Atoi(tura_id)
	if err != nil {
		log.Printf("Invalid tura ID: %v", err)
		http.Error(w, "Invalid tura ID", http.StatusBadRequest)
		return
	}

	sid := r.PathValue("sid")
	osid, err := primitive.ObjectIDFromHex(sid)
	if err != nil {
		log.Printf("Invalid sector ID: %v", err)
		http.Error(w, "Invalid sector ID", http.StatusBadRequest)
		return
	}

	coll := db.Collection("rja")
	if coll == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}

	collation := options.Collation{Locale: "pl", NumericOrdering: true, Strength: 1}
	sortOrder := bson.D{{Key: "sid", Value: 1}}
	opts := options.Find().SetSort(sortOrder).SetCollation(&collation)
	cur, err := coll.Find(r.Context(), bson.M{"sid": osid, "tura_id": turaID}, opts)
	if err != nil {
		log.Printf("Error finding buses for sector %s: %v", sid, err)
		http.Error(w, "Error finding buses for sector", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	type BusInfo struct {
		BusID       primitive.ObjectID `json:"id"`
		SraID       primitive.ObjectID `json:"sra_id"`
		SID         primitive.ObjectID `json:"sid"`
		SectorOrder int                `json:"sector_order"`
		Canceled    bool               `json:"canceled"`
		D1          string             `json:"d1"`
		D2          string             `json:"d2"`
		D3          string             `json:"d3"`
		A1          string             `json:"a1"`
		A2          string             `json:"a2"`
		A3          string             `json:"a3"`
	}
	buses := make([]BusInfo, 0)

	for cur.Next(r.Context()) {
		var rja db.RJA
		if err := cur.Decode(&rja); err != nil {
			log.Printf("Error decoding RJA document: %v", err)
			http.Error(w, "Error decoding RJA document", http.StatusInternalServerError)
			return
		}

		var sra db.SRA
		err = db.Collection("sra").FindOne(r.Context(), bson.M{"_id": rja.SraID}).Decode(&sra)
		if err != nil {
			log.Printf("Error finding SRA for RJA %s: %v", rja.ID.Hex(), err)
			http.Error(w, "Error finding SRA for RJA", http.StatusInternalServerError)
			return
		}

		var congregation db.Congregation
		err = db.Collection("congregations").FindOne(r.Context(), bson.M{"_id": sra.CongregationID, "tura": rja.SectorOrder}).Decode(&congregation)
		if err != nil {
			log.Printf("Error finding congregation for SRA %s: %v", sra.ID.Hex(), err)
		} else {
			buses = append(buses, BusInfo{
				BusID:       rja.ID,
				SraID:       rja.SraID,
				SID:         rja.SectorID,
				SectorOrder: rja.SectorOrder,
				Canceled:    sra.Canceled,
				D1:          helpers.FormatTime(rja.D1),
				D2:          helpers.FormatTime(rja.D2),
				D3:          helpers.FormatTime(rja.D3),
				A1:          helpers.FormatTime(rja.A1),
				A2:          helpers.FormatTime(rja.A2),
				A3:          helpers.FormatTime(rja.A3),
			})
		}
	}

	err = json.NewEncoder(w).Encode(buses)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
