package czw

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

/**
*	Inicjalizacja formularza wydawania - lista zborów aktualnej tury.
 */
func Get_Init(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tura := db.WhichTura(r.Context())
	if tura == nil {
		log.Println("czw.Init: no active tura")
		http.Error(w, "no active tura", http.StatusInternalServerError)
		return
	}

	coll := db.Collection("congregations")
	if coll == nil {
		log.Println("czw.Init: collection 'congregations' not found")
		http.Error(w, "Collection 'congregations' not found", http.StatusInternalServerError)
		return
	}

	cur, err := coll.Find(r.Context(), bson.M{"tura": tura.TID})
	if err != nil {
		log.Printf("czw.Init: error finding congregations: %v", err)
		http.Error(w, "Error finding congregations", http.StatusInternalServerError)
		return
	}

	congregations := make([]db.Congregation, 0)
	if err := cur.All(r.Context(), &congregations); err != nil {
		log.Printf("czw.Init: error decoding congregations: %v", err)
		http.Error(w, "Error decoding congregations", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(congregations); err != nil {
		log.Printf("czw.Init: error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
