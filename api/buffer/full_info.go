package buffer

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

/**
* Zwraca wszystkie statyczne informacje na temat bufora
* oraz listę przypisanych do niego autobusów
 */
func Get_FullInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("buffer.GetFullInfo called")

	terminal_name := r.PathValue("terminal_name")

	coll := db.Collection("terminals")
	if coll == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	// opis bufora
	var terminal db.Terminal
	err := coll.FindOne(r.Context(), bson.M{"name": terminal_name}).Decode(&terminal)
	if err != nil {
		log.Println("Error decoding terminal document:", err)
		http.Error(w, "Error decoding terminal document", http.StatusInternalServerError)
		return
	}

	// sektory współpracujące z buforem
	// buffer.Terminals

	err = json.NewEncoder(w).Encode(terminal)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
