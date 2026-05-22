package buffer

import (
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

	// tura := db.WhichTura(r.Context())
	terminal_id := r.PathValue("terminal_id")

	coll := db.Collection("terminals")
	if coll == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	// opis bufora
	var terminal db.Terminal
	err := coll.FindOne(r.Context(), bson.M{"_id": terminal_id}).Decode(&terminal)
	if err != nil {
		log.Println("Error decoding terminal document:", err)
		http.Error(w, "Error decoding terminal document", http.StatusInternalServerError)
		return
	}

	// sektory współpracujące z buforem
	// buffer.Terminals

}
