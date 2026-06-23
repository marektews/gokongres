package pk

import (
	"errors"
	"gokongres/db"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/**
* Sprawdzanie czy jest jeszcze wolny identyfikator do wykorzystania
*    :return: HTTP codes: 200 | 404
 */
func Get_IsFreePass(w http.ResponseWriter, r *http.Request) {
	collDepartments := db.Collection("departments")
	if collDepartments == nil {
		log.Println("Collection 'departments' not found")
		http.Error(w, "Collection 'departments' not found", http.StatusInternalServerError)
		return
	}

	collDepsPK := db.Collection("departments_pk")
	if collDepsPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	depName := r.PathValue("dep_name")
	tura := r.PathValue("tura")

	// pole "tura" w bazie jest typu int - PathValue zwraca string, więc bez konwersji
	// zapytanie nie dopasowałoby żadnego dokumentu (MongoDB porównuje z uwzględnieniem typu)
	turaInt, err := strconv.Atoi(tura)
	if err != nil {
		log.Printf("Invalid tura value %q: %v", tura, err)
		http.Error(w, "Invalid tura", http.StatusBadRequest)
		return
	}

	var department db.Department
	err = collDepartments.FindOne(r.Context(), bson.M{"name": depName, "tura": turaInt}).Decode(&department)
	if errors.Is(err, mongo.ErrNoDocuments) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("Error occurred while finding department:", err)
		http.Error(w, "Error occurred while finding department", http.StatusInternalServerError)
		return
	}

	count, err := collDepsPK.CountDocuments(r.Context(), bson.M{"department_id": department.ID})
	if err != nil {
		log.Println("Error occurred while counting documents:", err)
		http.Error(w, "Error occurred while counting documents", http.StatusInternalServerError)
		return
	}

	if count <= int64(department.Plimit) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
