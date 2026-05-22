package pk

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_FindPassID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type RequestData struct {
		DepartmentID string `json:"dep_id"`
		RegNum       string `json:"regnum"`
	}
	var requestData RequestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	depID, err := primitive.ObjectIDFromHex(requestData.DepartmentID)
	if err != nil {
		log.Printf("Error parsing department ID: %v", err)
		http.Error(w, "Invalid department ID", http.StatusBadRequest)
		return
	}

	collDepsPK := db.Collection("departments_pk")
	if collDepsPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	var departmentPK db.DepartmentPK
	filter := bson.M{
		"department_id": depID,
		"$or": bson.A{
			bson.M{"regnum1": requestData.RegNum},
			bson.M{"regnum2": requestData.RegNum},
			bson.M{"regnum3": requestData.RegNum},
		},
	}
	err = collDepsPK.FindOne(r.Context(), filter).Decode(&departmentPK)
	if err != nil {
		log.Printf("Error finding department: %v", err)
		http.Error(w, "Department not found", http.StatusNotFound)
		return
	}

	type ResponseData struct {
		PkID primitive.ObjectID `json:"pk_id"`
	}
	responseData := ResponseData{
		PkID: departmentPK.ID,
	}
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}
}
