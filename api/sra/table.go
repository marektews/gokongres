package sra

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"
	"slices"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_Table(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collCongr := db.Collection("congregations")
	if collCongr == nil {
		log.Println("Błąd: Nie można uzyskać kolekcji congregations z bazy danych")
		http.Error(w, "Błąd: Nie można uzyskać kolekcji congregations z bazy danych", http.StatusInternalServerError)
		return
	}

	collPilots := db.Collection("pilots")
	if collPilots == nil {
		log.Println("Błąd: Nie można uzyskać kolekcji pilots z bazy danych")
		http.Error(w, "Błąd: Nie można uzyskać kolekcji pilots z bazy danych", http.StatusInternalServerError)
		return
	}

	collSRA := db.Collection("sra")
	if collSRA == nil {
		log.Println("Błąd: Nie można uzyskać kolekcji SRA z bazy danych")
		http.Error(w, "Błąd: Nie można uzyskać kolekcji SRA z bazy danych", http.StatusInternalServerError)
		return
	}

	optsSRA := options.Find().SetSort(bson.M{"timestamp": -1}) // Sortowanie malejąco po dacie utworzenia

	curNoBus, err := collSRA.Find(r.Context(), bson.M{"nobus": bson.M{"$exists": true, "$ne": false}}, optsSRA)
	if err != nil {
		log.Println("Błąd: Nie można uzyskać danych z kolekcji SRA", err)
		http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji SRA", http.StatusInternalServerError)
		return
	}
	defer curNoBus.Close(r.Context())

	var sraNoBusTable []db.SRA_NoBus
	if err := curNoBus.All(r.Context(), &sraNoBusTable); err != nil {
		log.Println("Błąd: Nie można uzyskać danych z kolekcji SRA", err)
		http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji SRA", http.StatusInternalServerError)
		return
	}

	curHasBus, err := collSRA.Find(r.Context(), bson.M{"nobus": bson.M{"$exists": false}}, optsSRA)
	if err != nil {
		log.Println("Błąd: Nie można uzyskać danych z kolekcji SRA", err)
		http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji SRA", http.StatusInternalServerError)
		return
	}
	defer curHasBus.Close(r.Context())

	var sraBusTable []db.SRA
	if err := curHasBus.All(r.Context(), &sraBusTable); err != nil {
		log.Println("Błąd: Nie można uzyskać danych z kolekcji SRA", err)
		http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji SRA", http.StatusInternalServerError)
		return
	}

	type HasBusData struct {
		Id           string          `json:"id"`
		Info         string          `json:"info"`
		Timestamp    string          `json:"timestamp"`
		Congregation db.Congregation `json:"congregation"`
		Bus          BusData         `json:"bus"`
		Pilot1       db.Pilot        `json:"pilot1"`
		Pilot2       *db.Pilot       `json:"pilot2,omitempty"`
		Pilot3       *db.Pilot       `json:"pilot3,omitempty"`
	}

	type NoBusData struct {
		Id           string          `json:"id"`
		Congregation db.Congregation `json:"congregation"`
		NoBus        bool            `json:"nobus"`
		Timestamp    string          `json:"timestamp"`
	}

	type ResponseItem struct {
		NoBus  *NoBusData  `json:"nobus,omitempty"`
		HasBus *HasBusData `json:"hasbus,omitempty"`
	}

	safeString := func(s *string) string {
		if s != nil {
			return *s
		}
		return ""
	}
	safeInt := func(v *int) int {
		if v != nil {
			return *v
		}
		return 0
	}

	var response []ResponseItem
	for _, sra := range sraBusTable {
		var congregation db.Congregation
		err = collCongr.FindOne(r.Context(), bson.M{"_id": sra.CongregationID}).Decode(&congregation)
		if err != nil {
			log.Println("Błąd: Nie można uzyskać danych z kolekcji congregations", err)
			http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji congregations", http.StatusInternalServerError)
			return
		}

		PilotIDs := []primitive.ObjectID{sra.Pilot1ID}
		if sra.Pilot2ID != nil {
			PilotIDs = append(PilotIDs, *sra.Pilot2ID)
		}
		if sra.Pilot3ID != nil {
			PilotIDs = append(PilotIDs, *sra.Pilot3ID)
		}

		var pilots []db.Pilot
		pilotCur, err := collPilots.Find(r.Context(), bson.M{"_id": bson.M{"$in": PilotIDs}})
		if err != nil {
			log.Println("Błąd: Nie można uzyskać danych z kolekcji pilots", err)
			http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji pilots", http.StatusInternalServerError)
			return
		}
		if err = pilotCur.All(r.Context(), &pilots); err != nil {
			log.Println("Błąd: Nie można uzyskać danych z kolekcji pilots", err)
			http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji pilots", http.StatusInternalServerError)
			pilotCur.Close(r.Context())
			return
		}
		pilotCur.Close(r.Context())

		if len(pilots) > 0 {
			idx := slices.IndexFunc(pilots, func(p db.Pilot) bool {
				return p.ID == sra.Pilot1ID
			})

			responseItem := ResponseItem{
				HasBus: &HasBusData{
					Id:           sra.ID.Hex(),
					Info:         safeString(sra.Info),
					Timestamp:    sra.Timestamp.Time().Format("2006-01-02 15:04:05"),
					Congregation: congregation,
					Bus: BusData{
						Lp:               safeInt(sra.Lp),
						Prefix:           "", // sra.Prefix
						StaticIdentifier: safeString(sra.StaticIdentifier),
						Type:             sra.Bus.Type,
						Distance:         sra.Bus.Distance,
						ParkingMode:      sra.Bus.ParkingMode,
					},
					Pilot1: pilots[idx],
				},
			}
			if sra.Pilot2ID != nil {
				idx := slices.IndexFunc(pilots, func(p db.Pilot) bool {
					return p.ID == *sra.Pilot2ID
				})
				responseItem.HasBus.Pilot2 = &pilots[idx]
			}
			if sra.Pilot3ID != nil {
				idx := slices.IndexFunc(pilots, func(p db.Pilot) bool {
					return p.ID == *sra.Pilot3ID
				})
				responseItem.HasBus.Pilot3 = &pilots[idx]
			}

			response = append(response, responseItem)
		}
	}

	for _, sra := range sraNoBusTable {
		var congregation db.Congregation
		err = collCongr.FindOne(r.Context(), bson.M{"_id": sra.CongregationID}).Decode(&congregation)
		if err != nil {
			log.Println("Błąd: Nie można uzyskać danych z kolekcji congregations", err)
			http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji congregations", http.StatusInternalServerError)
			return
		}

		responseItem := ResponseItem{
			NoBus: &NoBusData{
				Id:           sra.ID.Hex(),
				Congregation: congregation,
				Timestamp:    sra.Timestamp.Time().Format("2006-01-02 15:04:05"),
			},
		}

		response = append(response, responseItem)
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Błąd: Nie można zakodować odpowiedzi JSON", err)
		return
	}
}
