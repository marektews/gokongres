package arrivals

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Set ustawia stan przyjazdu autokaru. Oczekuje JSON `{bus_id, state}`.
func Set(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		BusID string `json:"bus_id"`
		State bool   `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	busID, err := primitive.ObjectIDFromHex(req.BusID)
	if err != nil {
		log.Printf("arrivals.Set: invalid bus_id '%s': %v", req.BusID, err)
		http.Error(w, "Invalid bus_id", http.StatusBadRequest)
		return
	}

	if err := db.SetArrival(r.Context(), busID, req.State); err != nil {
		log.Printf("arrivals.Set: error setting arrival for '%s': %v", req.BusID, err)
		http.Error(w, "Error setting arrival", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
