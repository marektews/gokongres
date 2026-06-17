package rja

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Get_CongregationRJA zwraca rozkład jazdy autokarów danego zboru:
// zbór -> SRA (po congregation_id) -> RJA (po sra_id) -> nazwa sektora (po sector_id w terminalu).
func Get_CongregationRJA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("rja.GetCongregationRJA called")

	congregation_id := r.PathValue("congregation_id")
	congregationID, err := primitive.ObjectIDFromHex(congregation_id)
	if err != nil {
		log.Printf("Invalid congregation ID: %v", err)
		http.Error(w, "Invalid congregation ID", http.StatusBadRequest)
		return
	}

	deref := func(p *string) string {
		if p != nil {
			return *p
		}
		return ""
	}

	// mapa sid -> nazwa sektora (sektory są osadzone w terminalach)
	sectorNames := make(map[primitive.ObjectID]string)
	termColl := db.Collection("terminals")
	if termColl == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}
	termCur, err := termColl.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("Error finding terminals: %v", err)
		http.Error(w, "Error finding terminals", http.StatusInternalServerError)
		return
	}
	var terminals []db.Terminal
	if err := termCur.All(r.Context(), &terminals); err != nil {
		log.Printf("Error decoding terminals: %v", err)
		http.Error(w, "Error decoding terminals", http.StatusInternalServerError)
		return
	}
	for _, t := range terminals {
		for _, s := range t.Sectors {
			sectorNames[s.Sid] = s.Name
		}
	}

	// SRA tego zboru (mapa sraID -> SRA, by mieć static_identifier)
	sraColl := db.Collection("sra")
	if sraColl == nil {
		log.Println("Collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}
	sraCur, err := sraColl.Find(r.Context(), bson.M{"congregation_id": congregationID})
	if err != nil {
		log.Printf("Error finding SRA for congregation %s: %v", congregation_id, err)
		http.Error(w, "Error finding SRA for congregation", http.StatusInternalServerError)
		return
	}
	var sraList []db.SRA
	if err := sraCur.All(r.Context(), &sraList); err != nil {
		log.Printf("Error decoding SRA list: %v", err)
		http.Error(w, "Error decoding SRA list", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Sector string `json:"sector"`
		Ident  string `json:"ident"`
		D1     string `json:"d1"`
		D2     string `json:"d2"`
		D3     string `json:"d3"`
	}
	resp := make([]Response, 0)

	if len(sraList) == 0 {
		json.NewEncoder(w).Encode(resp)
		return
	}

	sraByID := make(map[primitive.ObjectID]db.SRA, len(sraList))
	sraIDs := make([]primitive.ObjectID, 0, len(sraList))
	for _, sra := range sraList {
		sraByID[sra.ID] = sra
		sraIDs = append(sraIDs, sra.ID)
	}

	// autokary (RJA) powiązane z tymi SRA
	rjaColl := db.Collection("rja")
	if rjaColl == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}
	rjaCur, err := rjaColl.Find(r.Context(), bson.M{"sra_id": bson.M{"$in": sraIDs}})
	if err != nil {
		log.Printf("Error finding RJA for congregation %s: %v", congregation_id, err)
		http.Error(w, "Error finding RJA for congregation", http.StatusInternalServerError)
		return
	}
	var buses []db.RJA
	if err := rjaCur.All(r.Context(), &buses); err != nil {
		log.Printf("Error decoding RJA list: %v", err)
		http.Error(w, "Error decoding RJA list", http.StatusInternalServerError)
		return
	}

	for _, bus := range buses {
		sra := sraByID[bus.SraID]
		sectorName := sectorNames[bus.SectorID]
		resp = append(resp, Response{
			Sector: sectorName,
			Ident:  db.CreateShortBusID(&sra, sectorName, bus.SectorOrder),
			D1:     deref(bus.D1),
			D2:     deref(bus.D2),
			D3:     deref(bus.D3),
		})
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
