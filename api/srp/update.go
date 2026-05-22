package srp

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
*	Aktualizacja przepustki parkingowej
 */
func Post_UpdatePassData(w http.ResponseWriter, r *http.Request) {
	type CarInfo struct {
		RegNum string `json:"regnum"`
		Lpg    bool   `json:"lpg"`
	}
	type RequestData struct {
		PassID               string   `json:"passid"`
		MobilityRestrictions bool     `json:"smr"`
		Car1                 CarInfo  `json:"car1"`
		Car2                 *CarInfo `json:"car2,omitempty"`
		Car3                 *CarInfo `json:"car3,omitempty"`
	}
	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	passIdent, err := primitive.ObjectIDFromHex(requestData.PassID)
	if err != nil {
		log.Printf("Invalid pass ID format: %v", err)
		http.Error(w, "Invalid pass ID format", http.StatusBadRequest)
		return
	}

	setFields := bson.M{
		"timestamp":             primitive.NewDateTimeFromTime(time.Now()),
		"mobility_restrictions": requestData.MobilityRestrictions,
		"car1.regnum":           requestData.Car1.RegNum,
		"car1.lpg":              requestData.Car1.Lpg,
	}
	unsetFields := bson.M{}

	if requestData.Car2 != nil && requestData.Car2.RegNum != "" {
		setFields["car2.regnum"] = requestData.Car2.RegNum
		setFields["car2.lpg"] = requestData.Car2.Lpg
	} else {
		unsetFields["car2"] = ""
	}

	if requestData.Car3 != nil && requestData.Car3.RegNum != "" {
		setFields["car3.regnum"] = requestData.Car3.RegNum
		setFields["car3.lpg"] = requestData.Car3.Lpg
	} else {
		unsetFields["car3"] = ""
	}

	update := bson.M{"$set": setFields}
	if len(unsetFields) > 0 {
		update["$unset"] = unsetFields
	}

	_, err = collSRP.UpdateOne(r.Context(), bson.M{"_id": passIdent}, update)
	if err != nil {
		log.Printf("Error updating SRP: %v", err)
		http.Error(w, "Error updating SRP", http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully updated SRP with ID %s", requestData.PassID)
}
