package pk

import (
	"errors"
	"fmt"
	"gokongres/db"
	"gokongres/helpers"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

/**
* Weryfikowanie identyfikatorów na parking księżycowy (działy)
*   :args: data=pk-<pass nr>-<regnum1>-<regnum2>-<regnum3>
*   :url: /api/pk/check?data=...
*   :return:
*       200 - może wjechać,
*       400 - brak wymaganego argumentu,
*       406 - argument ma niepoprawny format,
*       403 - nie może wjechać - identyfikator już użyty,
*       404 - brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu,
*       500 - błąd serwera - spróbuj ponownie później
 */
func Get_CheckPass(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		log.Println("Missing required argument 'data'")
		http.Error(w, "brak wymaganego argumentu", http.StatusBadRequest)
		return
	}

	// data=pk-<pass nr>-<regnum1>-<regnum2>-<regnum3>
	parts := strings.Split(data, "-")
	if len(parts) != 5 || parts[0] != "pk" {
		http.Error(w, "argument ma niepoprawny format", http.StatusNotAcceptable)
		return
	}

	passNr, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "argument ma niepoprawny format", http.StatusNotAcceptable)
		return
	}

	collDepsPK := db.Collection("departments_pk")
	if collDepsPK == nil {
		log.Println("Collection 'departments_pk' not found")
		http.Error(w, "błąd serwera - spróbuj ponownie później", http.StatusInternalServerError)
		return
	}

	var pkEntry db.DepartmentPK
	filter := bson.M{"pass_nr": passNr, "regnum1": parts[2]}
	err = collDepsPK.FindOne(r.Context(), filter).Decode(&pkEntry)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu", http.StatusNotFound)
			return
		}
		log.Println("Error occurred while finding PK entry:", err)
		http.Error(w, "błąd serwera - spróbuj ponownie później", http.StatusInternalServerError)
		return
	}

	// walidacja pozostałych numerów rejestracyjnych (jeśli zarejestrowane)
	if pkEntry.RegNum2 != nil && *pkEntry.RegNum2 != parts[3] {
		http.Error(w, "brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu", http.StatusNotFound)
		return
	}
	if pkEntry.RegNum3 != nil && *pkEntry.RegNum3 != parts[4] {
		http.Error(w, "brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu", http.StatusNotFound)
		return
	}

	// weryfikowanie użycia w aktywnym dniu kongresu
	dayField := helpers.GetActiveDayField()
	var used *time.Time
	switch dayField {
	case "d2":
		used = pkEntry.D2
	case "d3":
		used = pkEntry.D3
	default:
		used = pkEntry.D1
	}

	if used != nil {
		http.Error(w, fmt.Sprintf("identyfikator już użyty: %s", used.Format("2006-01-02 15:04:05")), http.StatusForbidden)
		return
	}

	_, err = collDepsPK.UpdateOne(
		r.Context(),
		bson.M{"_id": pkEntry.ID},
		bson.M{"$set": bson.M{dayField: time.Now()}},
	)
	if err != nil {
		log.Println("Error occurred while updating PK entry:", err)
		http.Error(w, "błąd serwera - spróbuj ponownie później", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("może wjechać"))
}
