package pk

import (
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

/**
* Weryfikowanie identyfikatorów na parking księżycowy
*   :args: OLD data=pk-<pass nr>-<regnum1>-<regnum2>-<regnum3>
*   :url: /api/pk/check/<pass nr>/<regnum1>/<regnum2>/<regnum3>
*   :return:
*       200 - może wjechać,
*       400 - brak wymaganego argumentu,
*       406 - argument ma niepoprawny format,
*       403 - nie może wjechać - identyfikator już użyty
*       404 - brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu,
*       500 - błąd serwera - spróbuj ponownie później
 */
func Get_CheckPass(w http.ResponseWriter, r *http.Request) {
	passNr := r.PathValue("pass_nr")
	regNum1 := r.PathValue("regnum1")
	regNum2 := r.PathValue("regnum2")
	regNum3 := r.PathValue("regnum3")

	// Walidacja argumentów
	if passNr == "" || regNum1 == "" || regNum2 == "" || regNum3 == "" {
		log.Println("Missing required arguments")
		http.Error(w, "Missing required arguments", http.StatusBadRequest)
		return
	}

	// dostęp do bazy
	collDepsPK := db.Collection("departments_pk")
	if collDepsPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "Collection 'departments_pk' not found", http.StatusInternalServerError)
		return
	}

	var pkEntry db.DepartmentPK
	filter := bson.M{
		"pass_nr": passNr,
		"regnum1": regNum1,
		"regnum2": regNum2,
		"regnum3": regNum3,
	}
	err := collDepsPK.FindOne(r.Context(), filter).Decode(&pkEntry)
	if err != nil {
		log.Println("Error occurred while finding PK entry:", err)
		http.Error(w, "brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu", http.StatusNotFound)
		return
	}

	// Sprawdzenie, czy identyfikator jest już użyty
	w.WriteHeader(http.StatusNotImplemented)
}
