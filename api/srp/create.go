package srp

import (
	"context"
	"encoding/json"
	"fmt"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Car struct {
	RegNum string `json:"regnum"`
	Lpg    bool   `json:"lpg"`
}
type RequestData struct {
	CongregationName     string `json:"congregation"`
	MobilityRestrictions bool   `json:"smr"`
	Car1                 Car    `json:"car1"`
	Car2                 *Car   `json:"car2,omitempty"`
	Car3                 *Car   `json:"car3,omitempty"`
}

func Post_Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var reqData RequestData
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	sr := collCongr.FindOne(r.Context(), bson.M{"name": reqData.CongregationName})
	if sr.Err() != nil {
		log.Printf("Error finding congregation: %v", sr.Err())
		http.Error(w, "Congregation not found", http.StatusNotFound)
		return
	}

	var congregation db.Congregation
	err = sr.Decode(&congregation)
	if err != nil {
		log.Printf("Error decoding congregation: %v", err)
		http.Error(w, "Error decoding congregation", http.StatusInternalServerError)
		return
	}

	// poszukiwanie nieużywanego jeszcze numeru identyfikatora
	usedNumbers := make([]int, congregation.Plimit)
	for i := range usedNumbers {
		usedNumbers[i] = i + 1
	}

	unusedNumbers, err := findUnusedNumbers(r.Context(), &reqData, congregation.ID, usedNumbers)
	if err != nil {
		log.Printf("Error finding unused numbers: %v", err)
		http.Error(w, "Error finding unused numbers", http.StatusInternalServerError)
		return
	}

	if len(unusedNumbers) == 0 {
		log.Printf("No unused number available for congregation %s", congregation.Name)
		http.Error(w, "No unused number available for this congregation", http.StatusConflict)
		return
	}

	// tworzenie nowego wpisu z nowym identyfikatorem
	srp := db.SRP{
		CongregationID:       congregation.ID,
		PassNr:               unusedNumbers[0],
		Timestamp:            primitive.NewDateTimeFromTime(time.Now()),
		MobilityRestrictions: reqData.MobilityRestrictions,
		Car1: db.CarInfo{
			RegNum: reqData.Car1.RegNum,
			Lpg:    reqData.Car1.Lpg,
		},
	}
	if reqData.Car2 != nil {
		srp.Car2 = &db.CarInfo{
			RegNum: reqData.Car2.RegNum,
			Lpg:    reqData.Car2.Lpg,
		}
	}
	if reqData.Car3 != nil {
		srp.Car3 = &db.CarInfo{
			RegNum: reqData.Car3.RegNum,
			Lpg:    reqData.Car3.Lpg,
		}
	}

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	insRes, err := collSRP.InsertOne(r.Context(), srp)
	if err != nil {
		log.Printf("Error inserting SRP entry: %v", err)
		http.Error(w, "Error inserting SRP entry", http.StatusInternalServerError)
		return
	}

	type ResponseData struct {
		PassID primitive.ObjectID `json:"passID"`
	}
	respData := ResponseData{
		PassID: insRes.InsertedID.(primitive.ObjectID),
	}
	err = json.NewEncoder(w).Encode(respData)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

/**
 * Funkcja pomocnicza do znajdowania nieużywanych numerów identyfikatorów oraz sprawdzania, czy podane numery rejestracyjne nie są już używane w innych wpisach SRP.
 * Przyjmuje kontekst, dane z żądania, ID zboru oraz listę używanych numerów identyfikatorów.
 * Zwraca listę nieużywanych numerów identyfikatorów lub błąd, jeśli wystąpił problem podczas przeszukiwania bazy danych.
 */
func findUnusedNumbers(ctx context.Context, reqData *RequestData, congregationID primitive.ObjectID, usedNumbers []int) ([]int, error) {
	coll := db.Collection("srp")
	if coll == nil {
		log.Println("Collection 'srp' not found")
		return nil, fmt.Errorf("collection 'srp' not found")
	}
	cur, err := coll.Find(ctx, bson.M{"congregation_id": congregationID})
	if err != nil {
		return nil, fmt.Errorf("error finding SRP entries: %v", err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var srp db.SRP
		if err := cur.Decode(&srp); err != nil {
			log.Printf("Error decoding SRP document: %v", err)
			continue
		}

		for i, v := range usedNumbers {
			if v == srp.PassNr {
				usedNumbers = append(usedNumbers[:i], usedNumbers[i+1:]...)
				break
			}
		}

		// testowanie czy pojazd nie występuje już na innym identyfikatorze
		if srp.Car1.RegNum == reqData.Car1.RegNum || (reqData.Car2 != nil && srp.Car1.RegNum == reqData.Car2.RegNum) || (reqData.Car3 != nil && srp.Car1.RegNum == reqData.Car3.RegNum) {
			return nil, fmt.Errorf("registration number %s is already used in another SRP entry", srp.Car1.RegNum)
		}
		if srp.Car2 != nil && (srp.Car2.RegNum == reqData.Car1.RegNum || (reqData.Car2 != nil && srp.Car2.RegNum == reqData.Car2.RegNum) || (reqData.Car3 != nil && srp.Car2.RegNum == reqData.Car3.RegNum)) {
			return nil, fmt.Errorf("registration number %s is already used in another SRP entry", srp.Car2.RegNum)
		}
		if srp.Car3 != nil && (srp.Car3.RegNum == reqData.Car1.RegNum || (reqData.Car2 != nil && srp.Car3.RegNum == reqData.Car2.RegNum) || (reqData.Car3 != nil && srp.Car3.RegNum == reqData.Car3.RegNum)) {
			return nil, fmt.Errorf("registration number %s is already used in another SRP entry", srp.Car3.RegNum)
		}
	}

	return usedNumbers, nil
}
