package pk

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type LoginRequest struct {
		Login    string `json:"login"`
		Password string `json:"passwd"`
	}
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("pk.GetLogin: Error decoding login request: %v", err)
		http.Error(w, "Error decoding login request: "+err.Error(), http.StatusBadRequest)
		return
	}

	coll := db.Collection("departments")
	if coll == nil {
		log.Print("pk.GetLogin: no departments collection in database")
		http.Error(w, "no departments collection in database", http.StatusInternalServerError)
		return
	}

	departmentID, err := primitive.ObjectIDFromHex(req.Login)
	if err != nil {
		log.Printf("pk.GetLogin: invalid login format: %v", err)
		http.Error(w, "Invalid login format", http.StatusBadRequest)
		return
	}

	result := coll.FindOne(r.Context(), bson.M{"_id": departmentID, "password": req.Password})
	if result.Err() != nil {
		log.Printf("pk.GetLogin: invalid credentials for user %s, error: %v", req.Login, result.Err())
		http.Error(w, "Invalid login or password", http.StatusForbidden)
		return
	}
	// Login successful jeśli tu doszło - nie musimy nic więcej zwracać
}
