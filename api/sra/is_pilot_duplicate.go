package sra

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

/**
* Sprawdzanie czy podane dane pilota nie są już przypisane do innego autokaru
*    :return: HTTP codes: 200, 400
 */
func Post_IsPilotDuplicate(w http.ResponseWriter, r *http.Request) {
	type RequestData struct {
		Phone db.Phone `json:"phone"`
	}
	var reqData RequestData
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.Printf("Error decoding request data: %v", err)
		http.Error(w, "invalid request data, JSON expected", http.StatusBadRequest)
		return
	}

	collPilots := db.Collection("pilots")
	if collPilots == nil {
		log.Println("Collection 'pilots' not found")
		http.Error(w, "Collection 'pilots' not found", http.StatusInternalServerError)
		return
	}

	phone := db.Phone{
		CountryCode: reqData.Phone.CountryCode,
		Number:      reqData.Phone.Number,
	}

	count, err := collPilots.CountDocuments(r.Context(), bson.M{"phone": phone})
	if err != nil {
		log.Printf("Error finding pilot: %v", err)
		http.Error(w, "Error finding pilot", http.StatusInternalServerError)
		return
	}

	log.Println("Is pilot duplicate:", count, phone)

	if count == 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
