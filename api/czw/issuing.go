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
*	Wydanie zastępczego identyfikatora parkingowego.
*	Kody: 200 (OK), 423 (identyfikator w użyciu), 404 (zbór nie znaleziony), 400/500 (błędy).
 */
func Post_Issuing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type RequestData struct {
		Zbor    string  `json:"zbor"`
		Phone   string  `json:"phone"`
		NrRej   string  `json:"nr_rej"`
		NrIdent flexInt `json:"nr_ident"`
	}
	var req RequestData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("czw.Issuing: error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collCzw := db.Collection("czw")
	if collCzw == nil {
		log.Println("czw.Issuing: collection 'czw' not found")
		http.Error(w, "Collection 'czw' not found", http.StatusInternalServerError)
		return
	}
	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("czw.Issuing: collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	// czy identyfikator jest już aktywnie używany (nie anulowany)?
	count, err := collCzw.CountDocuments(r.Context(), bson.M{"nr_ident": int(req.NrIdent), "cancellation": nil})
	if err != nil {
		log.Printf("czw.Issuing: error checking existing identifier: %v", err)
		http.Error(w, "Error checking identifier", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		w.WriteHeader(http.StatusLocked) // 423
		json.NewEncoder(w).Encode(map[string]string{"error": "in used"})
		return
	}

	// odnalezienie zboru po nazwie
	var congregation db.Congregation
	if err := collCongr.FindOne(r.Context(), bson.M{"name": req.Zbor}).Decode(&congregation); err != nil {
		log.Printf("czw.Issuing: congregation '%s' not found: %v", req.Zbor, err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "zbor " + req.Zbor + " not found"})
		return
	}

	czw := db.Czw{
		NrRej:          req.NrRej,
		Phone:          req.Phone,
		NrIdent:        int(req.NrIdent),
		CongregationID: congregation.ID,
		Issuing:        time.Now(),
	}
	if _, err := collCzw.InsertOne(r.Context(), czw); err != nil {
		log.Printf("czw.Issuing: error inserting czw entry: %v", err)
		http.Error(w, "Error issuing identifier", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{})
}
