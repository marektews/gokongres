package buffer

import (
	"encoding/json"
	"gokongres/db"
	"gokongres/helpers"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
* Zwraca wszystkie statyczne informacje na temat bufora
* oraz listę przypisanych do niego autobusów (przyjeżdżających w aktywnym dniu)
 */
func Get_FullInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("buffer.GetFullInfo called")

	terminal_name := r.PathValue("terminal_name")

	coll := db.Collection("terminals")
	if coll == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	// opis bufora
	var terminal db.Terminal
	err := coll.FindOne(r.Context(), bson.M{"name": terminal_name}).Decode(&terminal)
	if err != nil {
		log.Println("Error decoding terminal document:", err)
		http.Error(w, "Error decoding terminal document", http.StatusInternalServerError)
		return
	}

	// autobusy przypisane do bufora (przez sektory), które przyjeżdżają w aktywnym dniu
	arrived, err := arrivedRJAs(r.Context(), terminal)
	if err != nil {
		log.Printf("Error finding RJA documents for terminal '%s': %v", terminal_name, err)
		http.Error(w, "Error finding RJA documents", http.StatusInternalServerError)
		return
	}

	activeDay := helpers.GetActiveDay()
	collSRA := db.Collection("sra")
	collCong := db.Collection("congregations")

	type SectorInfo struct {
		Name string `json:"name"`
	}
	type SraInfo struct {
		Lp *int `json:"lp"`
	}
	type CongregationInfo struct {
		Ident string `json:"ident"`
		Name  string `json:"name"`
	}
	type BusInfo struct {
		ID           primitive.ObjectID `json:"id"`
		Arrive       string             `json:"arrive"`
		Departure    string             `json:"departure"`
		Sra          SraInfo            `json:"sra"`
		Sector       SectorInfo         `json:"sector"`
		Congregation CongregationInfo   `json:"congregation"`
	}
	type Response struct {
		TerminalID primitive.ObjectID `json:"bid"`
		Name       string             `json:"name"`
		Buses      []BusInfo          `json:"buses"` // inicjalizacja [] → nigdy null
	}

	resp := Response{
		TerminalID: terminal.ID,
		Name:       terminal.Name,
		Buses:      []BusInfo{},
	}

	for _, a := range arrived {
		var sra db.SRA
		if err := collSRA.FindOne(r.Context(), bson.M{"_id": a.RJA.SraID}).Decode(&sra); err != nil {
			log.Printf("FullInfo: SRA not found for rja %s: %v", a.RJA.ID.Hex(), err)
			continue // pomijamy autobus bez SRA
		}

		var cong db.Congregation
		if err := collCong.FindOne(r.Context(), bson.M{"_id": sra.CongregationID}).Decode(&cong); err != nil {
			log.Printf("FullInfo: congregation not found for sra (congregation_id %s): %v", sra.CongregationID.Hex(), err)
			continue // pomijamy autobus bez zboru
		}

		resp.Buses = append(resp.Buses, BusInfo{
			ID:        a.RJA.ID,
			Arrive:    a.RJA.ArriveByDay(activeDay),
			Departure: a.RJA.DepartureByDay(activeDay),
			Sra:       SraInfo{Lp: sra.Lp},
			Sector:    SectorInfo{Name: a.Sector.Name},
			Congregation: CongregationInfo{
				Ident: db.CreateShortBusID(&sra, a.Sector.Name, a.RJA.SectorOrder),
				Name:  cong.Name,
			},
		})
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
