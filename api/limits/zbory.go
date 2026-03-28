package limits

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// zborResponse to typ używany w odpowiedziach JSON
type zborResponse struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Name   string `json:"name"`
	Lang   string `json:"lang"`
	Plimit int    `json:"plimit"`
	Tura   int    `json:"tura"`
}

// GetZbory zwraca wszystkie zbiory (zbory) jako JSON.
func GetZbory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	zbory, err := db.GetAllZbory(r.Context())
	if err != nil {
		log.Printf("GetZbory: Error fetching zbory: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]zborResponse, 0, len(zbory))
	for _, z := range zbory {
		resp = append(resp, zborResponse{
			ID:     z.ID.Hex(),
			Number: z.Number,
			Name:   z.Name,
			Lang:   z.Lang,
			Plimit: z.Plimit,
			Tura:   z.Tura,
		})
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("GetZbory: Error encoding response: %v", err)
		return
	}
}

// SetZboryLimit request body
// {"zbor_id": "<hex>", "plimit": 123}
type setZborLimitReq struct {
	ZborID string `json:"zbor_id"`
	Plimit int    `json:"plimit"`
}

// SetZboryLimit aktualizuje limit dla wybranego zboru.
func SetZboryLimit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req setZborLimitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("SetZboryLimit: error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// jeżeli brak klienta bazy -> 404
	if err := db.UpdateZboryLimit(r.Context(), req.ZborID, req.Plimit); err != nil {
		// jeśli id jest nieprawidłowy
		if err == primitive.ErrInvalidHex {
			log.Printf("SetZboryLimit: invalid id - %s", req.ZborID)
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		} else {
			log.Printf("SetZboryLimit: error updating limit for zbor_id %s: %v", req.ZborID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
