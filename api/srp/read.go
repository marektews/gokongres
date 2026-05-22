package srp

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
* Odczyt stanu identyfikatora
 */
func Get_ReadPassData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	passID := r.PathValue("pass_id")

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	passIdent, err := primitive.ObjectIDFromHex(passID)
	if err != nil {
		log.Printf("Invalid pass ID format: %v", err)
		http.Error(w, "Invalid pass ID format", http.StatusBadRequest)
		return
	}

	var srp db.SRP
	err = collSRP.FindOne(r.Context(), bson.M{"_id": passIdent}).Decode(&srp)
	if err != nil {
		log.Printf("Error finding SRP: %v", err)
		http.Error(w, "SRP not found", http.StatusNotFound)
		return
	}

	type CarInfo struct {
		RegNum string `json:"regnum"`
		Lpg    bool   `json:"lpg"`
	}
	type ResponseData struct {
		PassID               string  `json:"passid"`
		PassNr               int     `json:"pass_nr"`
		MobilityRestrictions bool    `json:"smr"`
		Car1                 CarInfo `json:"car1"`
		Car2                 CarInfo `json:"car2"`
		Car3                 CarInfo `json:"car3"`
	}
	responseData := ResponseData{
		PassID:               srp.ID.Hex(),
		PassNr:               srp.PassNr,
		MobilityRestrictions: srp.MobilityRestrictions,
		Car1: CarInfo{
			RegNum: srp.Car1.RegNum,
			Lpg:    srp.Car1.Lpg,
		},
	}
	if srp.Car2 != nil {
		responseData.Car2 = CarInfo{
			RegNum: srp.Car2.RegNum,
			Lpg:    srp.Car2.Lpg,
		}
	}
	if srp.Car3 != nil {
		responseData.Car3 = CarInfo{
			RegNum: srp.Car3.RegNum,
			Lpg:    srp.Car3.Lpg,
		}
	}

	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
