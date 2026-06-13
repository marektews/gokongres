package monitoring

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Get_StatesRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collRJA := db.Collection("rja")
	if collRJA == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'rja' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'rja' jest nil", http.StatusInternalServerError)
		return
	}

	cur, err := collRJA.Find(r.Context(), bson.M{})
	if err != nil {
		log.Println("Błąd podczas pobierania dokumentów z kolekcji 'rja':", err)
		http.Error(w, "Błąd podczas pobierania dokumentów z kolekcji 'rja'", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var allRJA []db.RJA
	if err = cur.All(r.Context(), &allRJA); err != nil {
		log.Println("Błąd podczas odczytywania dokumentów z kolekcji 'rja':", err)
		http.Error(w, "Błąd podczas odczytywania dokumentów z kolekcji 'rja'", http.StatusInternalServerError)
		return
	}

	collSOA := db.Collection("soa")
	if collSOA == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'soa' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'soa' jest nil", http.StatusInternalServerError)
		return
	}

	type RJAState struct {
		Status    string `json:"status"`
		Timestamp string `json:"ts"`
	}
	res := make(map[primitive.ObjectID]RJAState)

	for _, rja := range allRJA {
		if !rja.WasArrived() {
			continue
		}

		var soa db.SOA
		if err := collSOA.FindOne(r.Context(), bson.M{"rja_id": rja.ID}).Decode(&soa); err != nil {
			continue // brak dokumentu SOA (autokar bez statusu) → pomijamy
		}

		last, ok := soa.Latest()
		if !ok {
			continue // dokument bez stanów → pomijamy
		}

		res[rja.ID] = RJAState{
			Status:    last.State,
			Timestamp: last.Ts.Format(time.DateTime),
		}
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("Błąd podczas kodowania odpowiedzi JSON:", err)
		return
	}
}
