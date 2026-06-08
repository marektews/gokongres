package buffer

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_States(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("buffer.GetStates called")

	terminal_name := r.PathValue("terminal_name")

	collTerm := db.Collection("terminals")
	if collTerm == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	collRJA := db.Collection("rja")
	if collRJA == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}

	// opis bufora
	var terminal db.Terminal
	err := collTerm.FindOne(r.Context(), bson.M{"name": terminal_name}).Decode(&terminal)
	if err != nil {
		log.Println("Error decoding terminal document:", err)
		http.Error(w, "Error decoding terminal document", http.StatusInternalServerError)
		return
	}

	// wszystkie autobusy przypisane do bufora (na podstawie rozkładu jazdy)
	cur, err := collRJA.Find(r.Context(), bson.M{"terminal_id": terminal.ID})
	if err != nil {
		log.Println("Error finding RJA documents:", err)
		http.Error(w, "Error finding RJA documents", http.StatusInternalServerError)
		return
	}

	var allRJA []db.RJA
	err = cur.All(r.Context(), &allRJA)
	if err != nil {
		log.Println("Error decoding RJA documents:", err)
		http.Error(w, "Error decoding RJA documents", http.StatusInternalServerError)
		return
	}

	type BufferState struct {
		Status    string `json:"status"`
		Timestamp int64  `json:"ts"`
	}

	type Response struct {
		TerminalID primitive.ObjectID `json:"bid"`
		States     []BufferState      `json:"states"`
	}

	resp := Response{
		TerminalID: terminal.ID,
	}

	for _, rja := range allRJA {
		if !rja.WasArrived() {
			continue
		}

		state := BufferState{
			Status:    "",                //rja.Status,
			Timestamp: time.Now().Unix(), // rja.Timestamp,
		}
		resp.States = append(resp.States, state)
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
