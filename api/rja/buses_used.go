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

	coll := db.Collection("rja")
	if coll == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}

	cur, err := coll.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("Error finding buses for tura %s: %v", tura_id, err)
		http.Error(w, "Error finding buses for tura", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var buses []db.RJA
	err = cur.All(r.Context(), &buses)
	if err != nil {
		log.Printf("Error decoding buses for tura %s: %v", tura_id, err)
		http.Error(w, "Error decoding buses for tura", http.StatusInternalServerError)
		return
	}

	type Response struct {
		SraID primitive.ObjectID `json:"sra_id"`
	}
	resp := make([]Response, 0)
	for _, bus := range buses {
		var sra db.SRA
		err := db.Collection("sra").FindOne(r.Context(), bson.M{"bus_id": bus.SraID}).Decode(&sra)
		if err != nil {
			log.Printf("Error finding SRA for bus %s: %v", bus.SraID.Hex(), err)
		} else {
			var congregation db.Congregation
			err = db.Collection("congregations").FindOne(r.Context(), bson.M{"_id": sra.CongregationID, "tura_id": tura_id}).Decode(&congregation)
			if err != nil {
				log.Printf("Error finding congregation for SRA %s: %v", sra.ID.Hex(), err)
			} else {
				resp = append(resp, Response{
					SraID: bus.SraID,
				})
			}
		}
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding response for tura %s: %v", tura_id, err)
		http.Error(w, "Error encoding response for tura", http.StatusInternalServerError)
		return
	}
}
