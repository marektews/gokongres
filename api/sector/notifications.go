package sector

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Notification_SendToSector(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "send-to-sector")
}

func Notification_ReadyToLeave(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "ready-to-leave")
}

func Notification_OnSector(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "on-sector")
}

func Notification_OnRoad(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "on-the-road")
}

func notification(w http.ResponseWriter, r *http.Request, status string) {
	w.Header().Set("Content-Type", "application/json")
	rjaID := r.PathValue("rja_id")

	coll := db.Collection("soa")
	if coll == nil {
		log.Println("Collection 'soa' not found")
		http.Error(w, "Collection 'soa' not found", http.StatusInternalServerError)
		return
	}

	objID, err := primitive.ObjectIDFromHex(rjaID)
	if err != nil {
		log.Printf("Invalid rja_id '%s': %v", rjaID, err)
		http.Error(w, "Invalid rja_id", http.StatusBadRequest)
		return
	}
	soa := db.SOA{RjaID: objID, Status: status, Timestamp: time.Now()}
	_, err = coll.InsertOne(r.Context(), soa)
	if err != nil {
		log.Printf("Error inserting SOA for rja_id '%s': %v", rjaID, err)
		http.Error(w, "Error inserting notification", http.StatusInternalServerError)
		return
	}

	type Response struct {
		RjaID     primitive.ObjectID `json:"rja_id"`
		Status    string             `json:"status"`
		Timestamp string             `json:"ts"`
	}
	resp := Response{RjaID: objID, Status: status, Timestamp: soa.Timestamp.Format("02.01.2006 15:04:05")}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding response for rja_id '%s': %v", rjaID, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
