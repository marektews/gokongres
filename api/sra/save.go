package sra

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
)

/**
*	Zapis zmodyfikowanych danych w module administracyjnym
 */
func Post_Save(w http.ResponseWriter, r *http.Request) {
	type RequestData struct {
		Id     string    `json:"id"`
		Info   string    `json:"info"`
		Bus    BusData   `json:"bus"`
		Pilot1 db.Pilot  `json:"pilot1"`
		Pilot2 *db.Pilot `json:"pilot2,omitempty"`
		Pilot3 *db.Pilot `json:"pilot3,omitempty"`
	}

	var req RequestData
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("decode request data error: %v", err)
		http.Error(w, "decode request data error", http.StatusInternalServerError)
		return
	}

	log.Printf("Request data: %+v", req)

	// TODO: dokończyć lub przerobić (dodać oddzielny edytor danych pilotów a tutaj tylko wybór z listy)
	w.WriteHeader(http.StatusNotImplemented)
}
