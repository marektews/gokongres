package srp

import (
	"errors"
	"fmt"
	"gokongres/api/ws"
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
* Weryfikowanie identyfikatorów na parking pod trybuną (Łazienkowska)
*   :args: data=<pass nr>-<regnum1>-<regnum2>-<regnum3>
*   :url: /api/srp/check?data=...
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

	// data=<pass nr>-<regnum1>-<regnum2>-<regnum3>
	parts := strings.Split(data, "-")
	if len(parts) != 4 {
		http.Error(w, "argument ma niepoprawny format", http.StatusNotAcceptable)
		return
	}

	passNr, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "argument ma niepoprawny format", http.StatusNotAcceptable)
		return
	}

	collSRP := db.Collection("srp")
	if collSRP == nil {
		log.Println("Collection 'srp' not found")
		http.Error(w, "błąd serwera - spróbuj ponownie później", http.StatusInternalServerError)
		return
	}

	var srp db.SRP
	filter := bson.M{"pass_nr": passNr, "car1.regnum": parts[1]}
	err = collSRP.FindOne(r.Context(), filter).Decode(&srp)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu", http.StatusNotFound)
			return
		}
		log.Println("Error occurred while finding SRP entry:", err)
		http.Error(w, "błąd serwera - spróbuj ponownie później", http.StatusInternalServerError)
		return
	}

	// walidacja pozostałych numerów rejestracyjnych (jeśli zarejestrowane)
	if srp.Car2 != nil && srp.Car2.RegNum != parts[2] {
		http.Error(w, "brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu", http.StatusNotFound)
		return
	}
	if srp.Car3 != nil && srp.Car3.RegNum != parts[3] {
		http.Error(w, "brak zarejestrowanego identyfikatora na ten numer rejestracyjny pojazdu", http.StatusNotFound)
		return
	}

	// weryfikowanie użycia w aktywnym dniu kongresu
	dayField := helpers.GetActiveDayField()
	var used *time.Time
	switch dayField {
	case "d2":
		used = srp.D2
	case "d3":
		used = srp.D3
	default:
		used = srp.D1
	}

	if used != nil {
		http.Error(w, fmt.Sprintf("identyfikator już użyty: %s", used.Format("2006-01-02 15:04:05")), http.StatusForbidden)
		return
	}

	now := time.Now()
	_, err = collSRP.UpdateOne(
		r.Context(),
		bson.M{"_id": srp.ID},
		bson.M{"$set": bson.M{dayField: now}},
	)
	if err != nil {
		log.Println("Error occurred while updating SRP entry:", err)
		http.Error(w, "błąd serwera - spróbuj ponownie później", http.StatusInternalServerError)
		return
	}

	ws.PublishParking("srp", srp.ID, srp.PassNr, true, now)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("może wjechać"))
}
