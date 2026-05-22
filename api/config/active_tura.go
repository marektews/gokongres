package config

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
)

func Get_ActiveTura(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	activeTura := db.WhichTura(r.Context())
	if activeTura == nil {
		log.Println("No active tura found")
		http.Error(w, "No active tura found", http.StatusNotFound)
		return
	}

	err := json.NewEncoder(w).Encode(activeTura)
	if err != nil {
		log.Println("Error encoding active tura:", err)
		http.Error(w, "Error encoding active tura", http.StatusInternalServerError)
		return
	}
}
