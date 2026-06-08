package rja

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_SraList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// która tura jest aktywna
	turaId := r.PathValue("tura_id")
	log.Printf("Getting SRA list for tura ID: %s", turaId)

	// pobieranie listy zborów które przypisane są do tej tury
	congregations, err := db.GetCongregationsForTura(r.Context(), turaId)
	if err != nil {
		log.Printf("Error getting congregations for tura ID: %s, error: %v", turaId, err)
		http.Error(w, "Error getting congregations for tura", http.StatusInternalServerError)
		return
	}

	// pobieranie z bazy danych listy SRA
	collSRA := db.Collection("sra")
	if collSRA == nil {
		log.Println("Collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}

	// pobieraj wszystkie aktywne SRA, które są przypisane do zborów z tej tury
	cur, err := collSRA.Find(r.Context(), bson.M{"bus": bson.M{"$exists": true}, "congregation_id": bson.M{"$in": db.GetCongregationIDs(congregations)}})
	if err != nil {
		log.Printf("Error finding SRA for tura ID: %v, error: %v", turaId, err)
		http.Error(w, "Error finding SRA", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var sraList []db.SRA
	err = cur.All(r.Context(), &sraList)
	if err != nil {
		log.Println("Error decoding SRA list:", err)
		http.Error(w, "Error decoding SRA list", http.StatusInternalServerError)
		return
	}
	log.Printf("SRA list: %+v", sraList)

	// budowanie listy SRA dla podanej tury
	type Response struct {
		SraID            primitive.ObjectID `json:"id"`
		CongregationID   primitive.ObjectID `json:"congregation_id"`
		Bus              db.Bus             `json:"bus"`
		Lp               *int               `json:"lp,omitempty"`
		Canceled         bool               `json:"canceled"`
		Prefix           string             `json:"prefix"`
		StaticIdentifier *string            `json:"static_identifier,omitempty"`
	}
	resp := make([]Response, 0)
	for _, sra := range sraList {
		resp = append(resp, Response{
			SraID:          sra.ID,
			CongregationID: sra.CongregationID,
			Bus:            sra.Bus,
			Lp:             sra.Lp,
			Canceled:       sra.Canceled,
			// Prefix:           sra.Prefix,
			StaticIdentifier: sra.StaticIdentifier,
		})
	}

	// zwracanie listy SRA jako JSON
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
