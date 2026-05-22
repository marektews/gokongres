package sector

import (
	"gokongres/db"
	"log"
	"net/http"
)

func Initialize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// sectorID := r.PathValue("sector_id")

	activeTura := db.WhichTura(r.Context())
	if activeTura == nil {
		log.Println("No active tura found")
		http.Error(w, "No active tura found", http.StatusNotFound)
		return
	}

	// opis sektora

}
