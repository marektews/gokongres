package arrivals

import (
	"encoding/json"
	"fmt"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type item struct {
	BusID    string `json:"bus_id"`
	Name     string `json:"name"`
	Arrived  bool   `json:"arrived"`
	Datetime string `json:"datetime"`
}

// Get_All zwraca listę dzisiejszych autokarów (z rozkładu RJA) wraz ze stanem przyjazdu.
func Get_All(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	result := make([]item, 0)

	activeTura := db.WhichTura(r.Context())
	if activeTura == nil {
		json.NewEncoder(w).Encode(result)
		return
	}

	collTerm := db.Collection("terminals")
	collRJA := db.Collection("rja")
	collSRA := db.Collection("sra")
	collCong := db.Collection("congregations")
	if collTerm == nil || collRJA == nil || collSRA == nil || collCong == nil {
		log.Println("arrivals: required collection not found")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// mapa sid.Hex() -> nazwa sektora (z osadzonych sektorów terminali)
	var terminals []db.Terminal
	curT, err := collTerm.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("arrivals: error finding terminals: %v", err)
		http.Error(w, "Error finding terminals", http.StatusInternalServerError)
		return
	}
	if err := curT.All(r.Context(), &terminals); err != nil {
		log.Printf("arrivals: error decoding terminals: %v", err)
		http.Error(w, "Error decoding terminals", http.StatusInternalServerError)
		return
	}
	sectorName := make(map[string]string)
	for _, t := range terminals {
		for _, s := range t.Sectors {
			sectorName[s.Sid.Hex()] = s.Name
		}
	}

	arrivals, err := db.GetArrivalMap(r.Context())
	if err != nil {
		log.Printf("arrivals: error reading arrival states: %v", err)
		http.Error(w, "Error reading arrival states", http.StatusInternalServerError)
		return
	}

	var rjas []db.RJA
	curR, err := collRJA.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("arrivals: error finding RJA: %v", err)
		http.Error(w, "Error finding RJA", http.StatusInternalServerError)
		return
	}
	if err := curR.All(r.Context(), &rjas); err != nil {
		log.Printf("arrivals: error decoding RJA: %v", err)
		http.Error(w, "Error decoding RJA", http.StatusInternalServerError)
		return
	}

	for _, rja := range rjas {
		if !rja.WasArrived() {
			continue
		}

		var sra db.SRA
		if err := collSRA.FindOne(r.Context(), bson.M{"_id": rja.SraID}).Decode(&sra); err != nil {
			continue // brak SRA → pomijamy
		}

		congFilter := bson.M{
			"$and": []bson.M{
				{"_id": sra.CongregationID},
				{"$or": []bson.M{{"tura": nil}, {"tura": activeTura.TID}}},
			},
		}
		var cong db.Congregation
		if err := collCong.FindOne(r.Context(), congFilter).Decode(&cong); err != nil {
			continue // zbór nie w aktywnej turze → pomijamy
		}

		ident := db.CreateShortBusID(&sra, sectorName[rja.SectorID.Hex()], rja.SectorOrder)
		name := fmt.Sprintf("%s - %s", ident, cong.Name)
		if sra.Lp != nil {
			name = fmt.Sprintf("%s - %s %d", ident, cong.Name, *sra.Lp)
		}

		busID := rja.ID.Hex()
		it := item{BusID: busID, Name: name}
		if a, ok := arrivals[busID]; ok && a.Arrived {
			it.Arrived = true
			it.Datetime = a.DateTime.Format(time.RFC3339)
		}
		result = append(result, it)
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("arrivals: error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
