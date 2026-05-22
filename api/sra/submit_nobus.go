package sra

import (
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Put_SubmitNoBus(w http.ResponseWriter, r *http.Request) {
	congregationName := r.PathValue("congregation_name")

	sra := db.SRA_NoBus{
		NoBus:     true,
		Timestamp: primitive.NewDateTimeFromTime(time.Now()),
	}

	// podłączenie zboru po nazwie
	var congregation db.Congregation
	err := db.Collection("congregations").FindOne(r.Context(), bson.M{"name": congregationName}).Decode(&congregation)
	if err != nil {
		log.Printf("Error finding congregation: %v", err)
		http.Error(w, "Error finding congregation", http.StatusBadRequest)
		return
	}
	sra.CongregationID = congregation.ID

	// zapis do bazy danych
	coll := db.Collection("sra")
	if coll == nil {
		log.Println("Collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}

	_, err = coll.InsertOne(r.Context(), sra)
	if err != nil {
		log.Printf("Error inserting SRA submission into database: %v", err)
		http.Error(w, "Error saving registration", http.StatusInternalServerError)
		return
	}
}
