package rja

import (
	"encoding/json"
	"gokongres/db"
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

	sector_id := r.PathValue("sector_id")
	sectorID, err := primitive.ObjectIDFromHex(sector_id)
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
	sortOrder := bson.D{{Key: "sector_order", Value: 1}}
	opts := options.Find().SetSort(sortOrder).SetCollation(&collation)
	cur, err := coll.Find(r.Context(), bson.M{"sector_id": sectorID}, opts)
	if err != nil {
		log.Printf("Error finding buses for sector %s: %v", sector_id, err)
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
		congregationFilter := bson.M{
			"$and": []bson.M{
				{"_id": sra.CongregationID},
				{"$or": []bson.M{{"tura": nil}, {"tura": turaID}}},
			},
		}
		err = db.Collection("congregations").FindOne(r.Context(), congregationFilter).Decode(&congregation)
		if err != nil {
			// zbór nie jest przypisany do tej tury, pomijamy ten RJA
		} else {
			buses = append(buses, BusInfo{
				BusID:       rja.ID,
				SraID:       rja.SraID,
				SID:         rja.SectorID,
				SectorOrder: rja.SectorOrder,
				Canceled:    sra.Canceled,
				D1:          *rja.D1,
				D2:          *rja.D2,
				D3:          *rja.D3,
				A1:          *rja.A1,
				A2:          *rja.A2,
				A3:          *rja.A3,
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
