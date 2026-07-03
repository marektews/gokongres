package monitoring

import (
	"encoding/json"
	"gokongres/db"
	"gokongres/helpers"
	"log"
	"net/http"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// passUsage to stan pojedynczego identyfikatora parkingowego w aktywnym dniu.
// Pola Zbor/ZborName wypełnia tylko SRP — numeracja pass_nr jest tam per zbór,
// więc numer bez kontekstu zboru byłby niejednoznaczny (PK grupuje po działach).
type passUsage struct {
	ID       string `json:"id"`
	Nr       int    `json:"nr"`
	Used     bool   `json:"used"`
	Ts       string `json:"ts,omitempty"`        // godzina wjazdu (tylko dla zajętych)
	Zbor     int    `json:"zbor,omitempty"`      // numer zboru
	ZborName string `json:"zbor_name,omitempty"` // nazwa zboru (do podpowiedzi)
}

// dayUsage wybiera znacznik użycia odpowiadający aktywnemu dniowi kongresu.
func dayUsage(d1, d2, d3 *time.Time, dayField string) *time.Time {
	switch dayField {
	case "d2":
		return d2
	case "d3":
		return d3
	default:
		return d1
	}
}

/**
* Stan użycia identyfikatorów na parkingu pod trybuną (SRP) w aktywnym dniu.
* Endpoint publiczny dla ekranu monitoringu.
*   :url: /api/monitoring/parking/srp
*   :return: {"total": N, "used": M, "passes": [{"id", "nr", "used", "ts"?}, ...]}
 */
func Get_ParkingSrp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	coll := db.Collection("srp")
	if coll == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'srp' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'srp' jest nil", http.StatusInternalServerError)
		return
	}

	cur, err := coll.Find(r.Context(), bson.M{})
	if err != nil {
		log.Println("Błąd podczas pobierania dokumentów z kolekcji 'srp':", err)
		http.Error(w, "Błąd podczas pobierania dokumentów z kolekcji 'srp'", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var allSRP []db.SRP
	if err = cur.All(r.Context(), &allSRP); err != nil {
		log.Println("Błąd podczas odczytywania dokumentów z kolekcji 'srp':", err)
		http.Error(w, "Błąd podczas odczytywania dokumentów z kolekcji 'srp'", http.StatusInternalServerError)
		return
	}

	// mapa zborów do rozwiązania numeracji per zbór (pass_nr nie jest globalny)
	collCong := db.Collection("congregations")
	if collCong == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'congregations' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'congregations' jest nil", http.StatusInternalServerError)
		return
	}

	curCong, err := collCong.Find(r.Context(), bson.M{})
	if err != nil {
		log.Println("Błąd podczas pobierania dokumentów z kolekcji 'congregations':", err)
		http.Error(w, "Błąd podczas pobierania dokumentów z kolekcji 'congregations'", http.StatusInternalServerError)
		return
	}
	defer curCong.Close(r.Context())

	var congregations []db.Congregation
	if err = curCong.All(r.Context(), &congregations); err != nil {
		log.Println("Błąd podczas odczytywania dokumentów z kolekcji 'congregations':", err)
		http.Error(w, "Błąd podczas odczytywania dokumentów z kolekcji 'congregations'", http.StatusInternalServerError)
		return
	}

	byID := make(map[primitive.ObjectID]db.Congregation, len(congregations))
	for _, cong := range congregations {
		byID[cong.ID] = cong
	}

	dayField := helpers.GetActiveDayField()

	usedCount := 0
	passes := make([]passUsage, 0, len(allSRP))
	for _, srp := range allSRP {
		p := passUsage{ID: srp.ID.Hex(), Nr: srp.PassNr}
		if cong, ok := byID[srp.CongregationID]; ok {
			p.Zbor = cong.Number
			p.ZborName = cong.Name
		}
		if used := dayUsage(srp.D1, srp.D2, srp.D3, dayField); used != nil {
			p.Used = true
			p.Ts = used.Local().Format(time.TimeOnly)
			usedCount++
		}
		passes = append(passes, p)
	}

	sort.Slice(passes, func(i, j int) bool {
		if passes[i].Zbor != passes[j].Zbor {
			return passes[i].Zbor < passes[j].Zbor
		}
		return passes[i].Nr < passes[j].Nr
	})

	err = json.NewEncoder(w).Encode(map[string]any{
		"total":  len(passes),
		"used":   usedCount,
		"passes": passes,
	})
	if err != nil {
		log.Println("Błąd podczas kodowania odpowiedzi JSON:", err)
	}
}
