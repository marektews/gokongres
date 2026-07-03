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
)

// passUsage to stan pojedynczego identyfikatora parkingowego w aktywnym dniu.
type passUsage struct {
	ID   string `json:"id"`
	Nr   int    `json:"nr"`
	Used bool   `json:"used"`
	Ts   string `json:"ts,omitempty"` // godzina wjazdu (tylko dla zajętych)
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

	dayField := helpers.GetActiveDayField()

	usedCount := 0
	passes := make([]passUsage, 0, len(allSRP))
	for _, srp := range allSRP {
		p := passUsage{ID: srp.ID.Hex(), Nr: srp.PassNr}
		if used := dayUsage(srp.D1, srp.D2, srp.D3, dayField); used != nil {
			p.Used = true
			p.Ts = used.Local().Format(time.TimeOnly)
			usedCount++
		}
		passes = append(passes, p)
	}

	sort.Slice(passes, func(i, j int) bool { return passes[i].Nr < passes[j].Nr })

	err = json.NewEncoder(w).Encode(map[string]any{
		"total":  len(passes),
		"used":   usedCount,
		"passes": passes,
	})
	if err != nil {
		log.Println("Błąd podczas kodowania odpowiedzi JSON:", err)
	}
}
