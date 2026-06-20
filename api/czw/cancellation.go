package czw

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

/**
*	Anulowanie aktywnego identyfikatora po numerze identyfikatora.
*	Kody: 200 (anulowano), 404 (identyfikator nie jest aktywnie używany), 400/500 (błędy).
 */
func Post_Cancellation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type RequestData struct {
		NrIdent flexInt `json:"nr_ident"`
	}
	var req RequestData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("czw.Cancellation: error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collCzw := db.Collection("czw")
	if collCzw == nil {
		log.Println("czw.Cancellation: collection 'czw' not found")
		http.Error(w, "Collection 'czw' not found", http.StatusInternalServerError)
		return
	}

	res, err := collCzw.UpdateOne(
		r.Context(),
		bson.M{"nr_ident": int(req.NrIdent), "cancellation": nil},
		bson.M{"$set": bson.M{"cancellation": time.Now()}},
	)
	if err != nil {
		log.Printf("czw.Cancellation: error updating czw entry: %v", err)
		http.Error(w, "Error cancelling identifier", http.StatusInternalServerError)
		return
	}

	if res.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{})
}
