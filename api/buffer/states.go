package buffer

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_States(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	terminal_name := r.PathValue("terminal_name")

	collTerm := db.Collection("terminals")
	if collTerm == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	soaColl := db.Collection("soa")
	if soaColl == nil {
		log.Println("Collection 'soa' not found")
		http.Error(w, "Collection 'soa' not found", http.StatusInternalServerError)
		return
	}

	// opis bufora
	var terminal db.Terminal
	err := collTerm.FindOne(r.Context(), bson.M{"name": terminal_name}).Decode(&terminal)
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

	type BufferState struct {
		Status    string `json:"status"`
		Timestamp string `json:"ts"`
	}

	type Response struct {
		TerminalID primitive.ObjectID     `json:"bid"`
		States     map[string]BufferState `json:"states"`
	}

	resp := Response{
		TerminalID: terminal.ID,
		States:     map[string]BufferState{}, // inicjalizacja → "{}" zamiast null w JSON
	}

	for _, a := range arrived {
		var soa db.SOA
		err = soaColl.FindOne(r.Context(), bson.M{"rja_id": a.RJA.ID}).Decode(&soa)
		if err != nil {
			log.Printf("Error finding SOA for rja_id '%s': %v", a.RJA.ID.Hex(), err)
			continue // brak dokumentu SOA → pomijamy ten autobus
		}

		last, ok := soa.Latest()
		if !ok {
			continue // dokument bez stanów → pomijamy
		}

		resp.States[a.RJA.ID.Hex()] = BufferState{
			Status:    last.State,
			Timestamp: last.Ts.Format("02.01.2006 15:04:05"),
		}
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
