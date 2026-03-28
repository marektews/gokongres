package limits

import (
	"encoding/json"
	"net/http"

	"gokongres/db"

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

// dzialResponse to typ używany w odpowiedziach JSON
type dzialResponse struct {
	ID     string `json:"id"`
	Lang   string `json:"lang"`
	Name   string `json:"name"`
	Plimit int    `json:"plimit"`
	Tura   int    `json:"tura"`
}

// GetZbory zwraca wszystkie zbiory (zbory) jako JSON.
func GetZbory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if db.Client() == nil {
		// bez bazy zwracamy pustą listę
		json.NewEncoder(w).Encode([]zborResponse{})
		return
	}
	ctx := r.Context()
	zbory, err := db.GetAllZbory(ctx)
	if err != nil {
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
	json.NewEncoder(w).Encode(resp)
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req setZborLimitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// jeżeli brak klienta bazy -> 404
	if db.Client() == nil {
		http.Error(w, "db not configured", http.StatusInternalServerError)
		return
	}
	if err := db.UpdateZboryLimit(r.Context(), req.ZborID, req.Plimit); err != nil {
		// jeśli id jest nieprawidłowy
		if err == primitive.ErrInvalidHex {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetDzialy zwraca wszystkie dzialy jako JSON.
func GetDzialy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if db.Client() == nil {
		json.NewEncoder(w).Encode([]dzialResponse{})
		return
	}
	ctx := r.Context()
	dzialy, err := db.GetAllDzialy(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]dzialResponse, 0, len(dzialy))
	for _, d := range dzialy {
		resp = append(resp, dzialResponse{
			ID:     d.ID.Hex(),
			Lang:   d.Lang,
			Name:   d.Name,
			Plimit: d.Plimit,
			Tura:   d.Tura,
		})
	}
	json.NewEncoder(w).Encode(resp)
}

// SetDzialyLimit request
// {"dzial_id": "<hex>", "plimit": 456}
type setDzialLimitReq struct {
	DzialID string `json:"dzial_id"`
	Plimit  int    `json:"plimit"`
}

// SetDzialyLimit aktualizuje limit dla dzialu
func SetDzialyLimit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req setDzialLimitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if db.Client() == nil {
		http.Error(w, "db not configured", http.StatusInternalServerError)
		return
	}
	if err := db.UpdateDzialyLimit(r.Context(), req.DzialID, req.Plimit); err != nil {
		if err == primitive.ErrInvalidHex {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
