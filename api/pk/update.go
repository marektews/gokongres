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

func Post_UpdatePassData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type RequestData struct {
		PkID    string  `json:"pkid"`
		RegNum1 string  `json:"regnum1"`
		RegNum2 *string `json:"regnum2,omitempty"`
		RegNum3 *string `json:"regnum3,omitempty"`
	}

	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pkID, err := primitive.ObjectIDFromHex(requestData.PkID)
	if err != nil {
		log.Printf("Invalid pkid format: %v", err)
		http.Error(w, "Invalid pkid format", http.StatusBadRequest)
		return
	}

	collDepsPK := db.Collection("departments_pk")
	if collDepsPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	setFields := bson.M{
		"timestamp": primitive.NewDateTimeFromTime(time.Now()),
		"regnum1":   requestData.RegNum1,
	}
	unsetFields := bson.M{}

	if requestData.RegNum2 != nil && *requestData.RegNum2 != "" {
		setFields["regnum2"] = *requestData.RegNum2
	} else {
		unsetFields["regnum2"] = ""
	}

	if requestData.RegNum3 != nil && *requestData.RegNum3 != "" {
		setFields["regnum3"] = *requestData.RegNum3
	} else {
		unsetFields["regnum3"] = ""
	}

	update := bson.M{"$set": setFields}
	if len(unsetFields) > 0 {
		update["$unset"] = unsetFields
	}

	_, err = collDepsPK.UpdateOne(r.Context(), bson.M{"_id": pkID}, update)
	if err != nil {
		log.Printf("Error updating departments_pk: %v", err)
		http.Error(w, "Error updating departments_pk", http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully updated departments_pk with ID: %s", requestData.PkID)
}
