package czw

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/**
*	Wyszukanie aktywnego identyfikatora po numerze identyfikatora lub numerze rejestracyjnym.
*	Kody: 200 (znaleziono), 404 (brak), 400/500 (błędy).
 */
func Post_Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type RequestData struct {
		NrIdent *flexInt `json:"nr_ident,omitempty"`
		NrRej   *string  `json:"nr_rej,omitempty"`
	}
	var req RequestData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("czw.Search: error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var filter bson.M
	switch {
	case req.NrIdent != nil:
		filter = bson.M{"nr_ident": int(*req.NrIdent), "cancellation": nil}
	case req.NrRej != nil:
		filter = bson.M{"nr_rej": *req.NrRej, "cancellation": nil}
	default:
		log.Println("czw.Search: neither nr_ident nor nr_rej provided")
		http.Error(w, "nr_ident or nr_rej required", http.StatusBadRequest)
		return
	}

	collCzw := db.Collection("czw")
	if collCzw == nil {
		log.Println("czw.Search: collection 'czw' not found")
		http.Error(w, "Collection 'czw' not found", http.StatusInternalServerError)
		return
	}

	var czw db.Czw
	err := collCzw.FindOne(r.Context(), filter).Decode(&czw)
	if err == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("czw.Search: error finding czw entry: %v", err)
		http.Error(w, "Error searching identifier", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(czw); err != nil {
		log.Printf("czw.Search: error encoding response: %v", err)
	}
}
