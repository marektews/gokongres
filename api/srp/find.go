package srp

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

func Post_FindPassID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type RequestData struct {
		CongregationName string `json:"congregation"`
		RegNum           string `json:"regnum"`
	}
	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
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

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "Collection 'srp' not found", http.StatusInternalServerError)
		return
	}

	var congregation db.Congregation
	err = collCongr.FindOne(r.Context(), bson.M{"name": requestData.CongregationName}).Decode(&congregation)
	if err != nil {
		log.Printf("Error finding congregation: %v", err)
		http.Error(w, "Congregation not found", http.StatusNotFound)
		return
	}

	var srp db.SRP
	filter := bson.M{
		"congregation_id": congregation.ID,
		"$or": bson.A{
			bson.M{"car1.regnum": requestData.RegNum},
			bson.M{"car2.regnum": requestData.RegNum},
			bson.M{"car3.regnum": requestData.RegNum},
		},
	}
	err = collSRP.FindOne(r.Context(), filter).Decode(&srp)
	if err != nil {
		log.Printf("Error finding SRP: %v", err)
		http.Error(w, "SRP not found", http.StatusNotFound)
		return
	}

	type ResponseData struct {
		PassID string `json:"pass_id"`
	}
	respData := ResponseData{
		PassID: srp.ID.Hex(),
	}
	err = json.NewEncoder(w).Encode(respData)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
