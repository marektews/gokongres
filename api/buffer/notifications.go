package buffer

import (
	"encoding/json"
	"gokongres/api/ws"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_NoBusNotification(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "no-bus")
}

func Get_InBufferNotification(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "in-buffer")
}

func Get_SecondCircleNotification(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "second-circle")
}

func Get_SendToSectorNotification(w http.ResponseWriter, r *http.Request) {
	notification(w, r, "send-to-sector")
}

func notification(w http.ResponseWriter, r *http.Request, status string) {
	w.Header().Set("Content-Type", "application/json")
	rjaID := r.PathValue("rja_id")

	objID, err := primitive.ObjectIDFromHex(rjaID)
	if err != nil {
		log.Printf("Invalid rja_id '%s': %v", rjaID, err)
		http.Error(w, "Invalid rja_id", http.StatusBadRequest)
		return
	}

	ts, err := db.PushSOAState(r.Context(), objID, status)
	if err != nil {
		log.Printf("Error pushing SOA state for rja_id '%s': %v", rjaID, err)
		http.Error(w, "Error inserting notification", http.StatusInternalServerError)
		return
	}

	// rozgłoszenie zmiany do pozostałych ekranów obserwujących ten sektor
	ws.PublishState(r.Context(), objID, status, ts)

	type Response struct {
		RjaID     primitive.ObjectID `json:"rja_id"`
		Status    string             `json:"status"`
		Timestamp string             `json:"ts"`
	}
	resp := Response{RjaID: objID, Status: status, Timestamp: ts.Format("02.01.2006 15:04:05")}
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding response for rja_id '%s': %v", rjaID, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
