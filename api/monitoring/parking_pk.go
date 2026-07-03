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

// depUsage to stan identyfikatorów jednego działu na parkingu księżycowym.
type depUsage struct {
	Dep    string      `json:"dep"`
	Total  int         `json:"total"`
	Used   int         `json:"used"`
	Passes []passUsage `json:"passes"`
}

/**
* Stan użycia identyfikatorów na parkingu księżycowym (działy) w aktywnym dniu,
* pogrupowany po działach aktywnej tury. Endpoint publiczny dla ekranu monitoringu.
*   :url: /api/monitoring/parking/pk
*   :return: {"total": N, "used": M, "groups": [{"dep", "total", "used", "passes": [...]}, ...]}
 */
func Get_ParkingPk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collDeps := db.Collection("departments")
	if collDeps == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'departments' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'departments' jest nil", http.StatusInternalServerError)
		return
	}

	// działy zawężone do aktywnej tury; bez aktywnej tury pokazujemy wszystkie
	depFilter := bson.M{}
	if tura := db.WhichTura(r.Context()); tura != nil {
		depFilter["tura"] = tura.TID
	}

	curDeps, err := collDeps.Find(r.Context(), depFilter)
	if err != nil {
		log.Println("Błąd podczas pobierania dokumentów z kolekcji 'departments':", err)
		http.Error(w, "Błąd podczas pobierania dokumentów z kolekcji 'departments'", http.StatusInternalServerError)
		return
	}
	defer curDeps.Close(r.Context())

	var deps []db.Department
	if err = curDeps.All(r.Context(), &deps); err != nil {
		log.Println("Błąd podczas odczytywania dokumentów z kolekcji 'departments':", err)
		http.Error(w, "Błąd podczas odczytywania dokumentów z kolekcji 'departments'", http.StatusInternalServerError)
		return
	}

	collPK := db.Collection("departments_pk")
	if collPK == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'departments_pk' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'departments_pk' jest nil", http.StatusInternalServerError)
		return
	}

	curPK, err := collPK.Find(r.Context(), bson.M{})
	if err != nil {
		log.Println("Błąd podczas pobierania dokumentów z kolekcji 'departments_pk':", err)
		http.Error(w, "Błąd podczas pobierania dokumentów z kolekcji 'departments_pk'", http.StatusInternalServerError)
		return
	}
	defer curPK.Close(r.Context())

	var allPK []db.DepartmentPK
	if err = curPK.All(r.Context(), &allPK); err != nil {
		log.Println("Błąd podczas odczytywania dokumentów z kolekcji 'departments_pk':", err)
		http.Error(w, "Błąd podczas odczytywania dokumentów z kolekcji 'departments_pk'", http.StatusInternalServerError)
		return
	}

	dayField := helpers.GetActiveDayField()

	// grupowanie po działach; identyfikatory działów spoza filtra tury pomijamy
	byDep := make(map[primitive.ObjectID]*depUsage, len(deps))
	for _, dep := range deps {
		byDep[dep.ID] = &depUsage{Dep: dep.Name, Passes: []passUsage{}}
	}

	totalCount, usedCount := 0, 0
	for _, pk := range allPK {
		group, ok := byDep[pk.DepartmentID]
		if !ok {
			continue
		}

		p := passUsage{ID: pk.ID.Hex(), Nr: pk.PassNr}
		if used := dayUsage(pk.D1, pk.D2, pk.D3, dayField); used != nil {
			p.Used = true
			p.Ts = used.Local().Format(time.TimeOnly)
			group.Used++
			usedCount++
		}
		group.Passes = append(group.Passes, p)
		group.Total++
		totalCount++
	}

	groups := make([]depUsage, 0, len(byDep))
	for _, group := range byDep {
		sort.Slice(group.Passes, func(i, j int) bool { return group.Passes[i].Nr < group.Passes[j].Nr })
		groups = append(groups, *group)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Dep < groups[j].Dep })

	err = json.NewEncoder(w).Encode(map[string]any{
		"total":  totalCount,
		"used":   usedCount,
		"groups": groups,
	})
	if err != nil {
		log.Println("Błąd podczas kodowania odpowiedzi JSON:", err)
	}
}
