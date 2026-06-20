package sra

import (
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
*	Usunięcie wpisu SRA wraz z powiązanymi danymi pilotów (moduł ADMSRA).
*	Obsługuje zarówno wpisy z autokarem (z pilotami), jak i wpisy "nobus" (bez pilotów).
 */
func Get_Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sraIDHex := r.PathValue("sra_id")

	sraID, err := primitive.ObjectIDFromHex(sraIDHex)
	if err != nil {
		log.Printf("sra.Delete: invalid SRA ID '%s': %v", sraIDHex, err)
		http.Error(w, "Invalid SRA ID", http.StatusBadRequest)
		return
	}

	collSRA := db.Collection("sra")
	if collSRA == nil {
		log.Println("sra.Delete: collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}
	collPilots := db.Collection("pilots")
	if collPilots == nil {
		log.Println("sra.Delete: collection 'pilots' not found")
		http.Error(w, "Collection 'pilots' not found", http.StatusInternalServerError)
		return
	}

	// wczytanie wpisu, aby poznać powiązanych pilotów
	var sra db.SRA
	err = collSRA.FindOne(r.Context(), bson.M{"_id": sraID}).Decode(&sra)
	if err != nil {
		log.Printf("sra.Delete: SRA '%s' not found: %v", sraIDHex, err)
		http.Error(w, "SRA not found", http.StatusNotFound)
		return
	}

	// usunięcie powiązanych pilotów (wpisy "nobus" nie mają pilotów)
	pilotIDs := []primitive.ObjectID{}
	if !sra.Pilot1ID.IsZero() {
		pilotIDs = append(pilotIDs, sra.Pilot1ID)
	}
	if sra.Pilot2ID != nil {
		pilotIDs = append(pilotIDs, *sra.Pilot2ID)
	}
	if sra.Pilot3ID != nil {
		pilotIDs = append(pilotIDs, *sra.Pilot3ID)
	}
	if len(pilotIDs) > 0 {
		if _, err := collPilots.DeleteMany(r.Context(), bson.M{"_id": bson.M{"$in": pilotIDs}}); err != nil {
			log.Printf("sra.Delete: error deleting pilots for SRA '%s': %v", sraIDHex, err)
			http.Error(w, "Error deleting pilots", http.StatusInternalServerError)
			return
		}
	}

	// usunięcie samego wpisu SRA
	res, err := collSRA.DeleteOne(r.Context(), bson.M{"_id": sraID})
	if err != nil {
		log.Printf("sra.Delete: error deleting SRA '%s': %v", sraIDHex, err)
		http.Error(w, "Error deleting SRA", http.StatusInternalServerError)
		return
	}
	if res.DeletedCount == 0 {
		http.Error(w, "SRA not found", http.StatusNotFound)
		return
	}

	log.Printf("sra.Delete: successfully deleted SRA '%s'", sraIDHex)
}
