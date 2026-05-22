package srp

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_AllList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	coll := db.Collection("srp")
	if coll == nil {
		log.Printf("Error: Database collection 'srp' not found")
		http.Error(w, "Database collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	cur, err := coll.Find(r.Context(), bson.M{})
	if err != nil {
		log.Printf("Error fetching srp: %v", err)
		http.Error(w, "Error fetching srp", http.StatusInternalServerError)
		return
	}

	var srps []db.SRP
	if err = cur.All(r.Context(), &srps); err != nil {
		log.Printf("Error decoding srp: %v", err)
		http.Error(w, "Error decoding srp", http.StatusInternalServerError)
		return
	}

	type CarInfo struct {
		RegNum string `json:"regnum"`
		Lpg    bool   `json:"lpg"`
	}
	type Response struct {
		ID                   primitive.ObjectID `json:"id"`
		CongregationID       primitive.ObjectID `json:"congregation_id"`
		PassNr               int                `json:"pass_nr"`
		MobilityRestrictions bool               `json:"smr"`
		Car1                 CarInfo            `json:"car1"`
		Car2                 *CarInfo           `json:"car2,omitempty"`
		Car3                 *CarInfo           `json:"car3,omitempty"`
	}
	resp := make([]Response, 0)

	for _, srp := range srps {
		respItem := Response{
			ID:                   srp.ID,
			CongregationID:       srp.CongregationID,
			PassNr:               srp.PassNr,
			MobilityRestrictions: srp.MobilityRestrictions,
			Car1: CarInfo{
				RegNum: srp.Car1.RegNum,
				Lpg:    srp.Car1.Lpg,
			},
		}
		if srp.Car2 != nil {
			respItem.Car2 = &CarInfo{
				RegNum: srp.Car2.RegNum,
				Lpg:    srp.Car2.Lpg,
			}
		}
		if srp.Car3 != nil {
			respItem.Car3 = &CarInfo{
				RegNum: srp.Car3.RegNum,
				Lpg:    srp.Car3.Lpg,
			}
		}

		resp = append(resp, respItem)
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
