package arrivals

import (
	"encoding/json"
	"net/http"

	"gokongres/db"
)

// GetAll zwraca listę wszystkich arrivals w formacie JSON.
// Jeśli dostępny jest klient MongoDB, pobiera z bazy, w przeciwnym razie zwraca pamięciowy store.
func GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if db.Client() != nil {
		ctx := r.Context()
		docs, err := db.GetAllArrivals(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(docs)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(arrivalsList)
}
