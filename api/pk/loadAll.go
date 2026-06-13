package pk

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_LoadAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	coll := db.Collection("departments_pk")
	if coll == nil {
		log.Printf("Error: Database collection 'departments_pk' not found")
		http.Error(w, "Database collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	cur, err := coll.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("Error fetching department PKs: %v", err)
		http.Error(w, "Error fetching department PKs", http.StatusInternalServerError)
		return
	}

	var pks []db.DepartmentPK
	if err = cur.All(r.Context(), &pks); err != nil {
		log.Printf("Error decoding department PKs: %v", err)
		http.Error(w, "Error decoding department PKs", http.StatusInternalServerError)
		return
	}

	type Response struct {
		ID           primitive.ObjectID `json:"id"`
		DepartmentID primitive.ObjectID `json:"dep_id"`
		PassNr       int                `json:"pass_nr"`
		Regnum1      string             `json:"regnum1"`
		Regnum2      *string            `json:"regnum2,omitempty"`
		Regnum3      *string            `json:"regnum3,omitempty"`
		Registered   string             `json:"registered"`
	}
	resp := make([]Response, 0)

	for _, pk := range pks {
		resp = append(resp, Response{
			ID:           pk.ID,
			DepartmentID: pk.DepartmentID,
			PassNr:       pk.PassNr,
			Regnum1:      pk.RegNum1,
			Regnum2:      pk.RegNum2,
			Regnum3:      pk.RegNum3,
			Registered:   pk.Registered.Time().Format(time.DateTime)[:16], // Formatowanie daty do formatu "YYYY-MM-DD HH:MM" (obcinanie sekund)
		})
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding department PKs: %v", err)
		http.Error(w, "Error encoding department PKs", http.StatusInternalServerError)
		return
	}
}
