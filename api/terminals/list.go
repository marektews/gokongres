package terminals

import (
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_AllList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("terminals.GetAllList called")

	coll := db.Collection("buffers")
	if coll == nil {
		log.Println("Collection 'buffers' not found")
		http.Error(w, "Collection 'buffers' not found", http.StatusInternalServerError)
		return
	}

	collation := options.Collation{Locale: "pl", NumericOrdering: true, Strength: 1}
	sortOrder := bson.D{{Key: "name", Value: 1}}
	opts := options.Find().SetSort(sortOrder).SetCollation(&collation)

	cur, err := coll.Find(r.Context(), bson.M{}, opts)
	if err != nil {
		log.Printf("Error finding terminals: %v", err)
		http.Error(w, "Error finding terminals", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	w.WriteHeader(http.StatusNotImplemented)
}
