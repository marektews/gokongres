package pk

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_LoadAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	coll := db.Collection("department_pk")
	if coll == nil {
		log.Printf("Error: Database collection 'department_pk' not found")
		http.Error(w, "Database collection 'department_pk' not found", http.StatusInternalServerError)
		return
	}

	cur, err := coll.Find(r.Context(), nil)
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
		ID           primitive.ObjectID `bson:"_id,omitempty"`
		DepartmentID primitive.ObjectID `bson:"dzial_id"`
		PassNr       int                `bson:"pass_nr"`
		Regnum1      string             `bson:"regnum1"`
		Regnum2      *string            `bson:"regnum2,omitempty"`
		Regnum3      *string            `bson:"regnum3,omitempty"`
		Registered   string             `bson:"registered"`
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
