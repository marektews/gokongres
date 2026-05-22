package terminals

import (
	"gokongres/db"
	"log"
	"net/http"
)

func Get_FullInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("terminals.GetFullInfo called")

	// terminal_id := r.PathValue("terminal_id")

	coll := db.Collection("terminals")
	if coll == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	// opts := options.Find().SetSort(bson.D{{"name", 1}})
	// coll.Find(r.Context(), bson.M{"tid": tid}, opts)

	w.WriteHeader(http.StatusNotImplemented)
}
