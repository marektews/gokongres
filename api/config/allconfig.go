package config

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
)

func Get_AllConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cfg, err := db.GetConstConfig()
	if err != nil {
		log.Printf("Get_AllConfig: get config error: %v", err)
		http.Error(w, "get config error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(cfg)
	if err != nil {
		log.Printf("Get_AllConfig: error encoding tury to JSON: %v", err)
		http.Error(w, "error encoding tury to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
