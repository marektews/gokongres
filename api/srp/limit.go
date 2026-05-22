package srp

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func Get_UsingLimit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	congregationName := r.PathValue("congregation_name")

	var congregation db.Congregation
	err := collCongr.FindOne(r.Context(), bson.M{"name": congregationName}).Decode(&congregation)
	if err != nil {
		log.Println("Error occurred while finding congregation:", err)
		http.Error(w, "Error occurred while finding congregation", http.StatusInternalServerError)
		return
	}

	count, err := collSRP.CountDocuments(r.Context(), bson.M{"zbor_id": congregation.ID})
	if err != nil {
		log.Println("Error occurred while counting SRP documents:", err)
		http.Error(w, "Error occurred while counting SRP documents", http.StatusInternalServerError)
		return
	}

	type RespData struct {
		Plimit int   `json:"plimit"`
		Used   int64 `json:"used"`
	}
	respData := RespData{
		Plimit: congregation.Plimit,
		Used:   count,
	}
	err = json.NewEncoder(w).Encode(respData)
	if err != nil {
		log.Println("Error sending response:", err)
	}
}

/**
*	Prośba o zmianę limitu pojazdów
 */
func Post_RequestNewLimit(w http.ResponseWriter, r *http.Request) {
	type RequestData struct {
		CongregationName string `json:"congregation"`
		Plimit           int    `json:"plimit"`
		Reason           string `json:"reason"`
	}
	var reqData RequestData
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.Println("Error decoding request data:", err)
		http.Error(w, "Error decoding request data", http.StatusInternalServerError)
		return
	}

	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("Collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	setFields := bson.M{
		"limitRequest.plimit": reqData.Plimit,
		"limitRequest.reason": reqData.Reason,
	}
	update := bson.M{"$set": setFields}

	_, err = collCongr.UpdateOne(r.Context(), bson.M{"name": reqData.CongregationName}, update)
	if err != nil {
		log.Printf("Error updating SRP: %v", err)
		http.Error(w, "Error updating SRP", http.StatusInternalServerError)
		return
	}
}
