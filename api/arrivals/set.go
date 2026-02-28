package arrivals

import (
	"encoding/json"
	"net/http"

	"gokongres/db"
)

// Set dodaje nowy arrival. Oczekuje JSON body z polem `message`.
func Set(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var a Arrival
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If MongoDB client is configured, persist there.
	if db.Client() != nil {
		doc, err := db.InsertArrival(r.Context(), a.Message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(doc)
		return
	}

	mu.Lock()
	a.ID = nextID
	nextID++
	arrivalsList = append(arrivalsList, a)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}
