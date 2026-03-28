package limits

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// dzialResponse to typ używany w odpowiedziach JSON
type dzialResponse struct {
	ID     primitive.ObjectID `json:"id"`
	Lang   string             `json:"lang"`
	Name   string             `json:"name"`
	Plimit int                `json:"plimit"`
	TuraID primitive.ObjectID `json:"tura_id"`
}

// GetDzialy zwraca wszystkie dzialy jako JSON.
func GetDzialy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db.Client() == nil {
		err := json.NewEncoder(w).Encode([]dzialResponse{})
		if err != nil {
			log.Printf("GetDzialy: Error encoding response: %v", err)
			http.Error(w, "error encoding response", http.StatusInternalServerError)
		}
		return
	}

	dzialy, err := db.GetAllDepartments(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := make([]dzialResponse, 0, len(dzialy))
	for _, d := range dzialy {
		resp = append(resp, dzialResponse{
			ID:     d.ID,
			Lang:   d.Lang,
			Name:   d.Name,
			Plimit: d.Plimit,
			TuraID: d.TuraID,
		})
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("GetDzialy: Error encoding response: %v", err)
		return
	}
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

	var req setDzialLimitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("SetDzialyLimit: Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateDepartmentLimit(r.Context(), req.DzialID, req.Plimit); err != nil {
		if err == primitive.ErrInvalidHex {
			log.Printf("SetDzialyLimit: invalid id - %s", req.DzialID)
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		} else {
			log.Printf("SetDzialyLimit: Error updating limit for dzial.id %s: %v", req.DzialID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
