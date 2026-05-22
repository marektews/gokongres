package sra

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

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
	cur, err := collSRA.Find(r.Context(), bson.M{}, optsSRA)
	if err != nil {
		log.Println("Błąd: Nie można uzyskać danych z kolekcji SRA", err)
		http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji SRA", http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var sraTableData []db.SRA
	if err := cur.All(r.Context(), &sraTableData); err != nil {
		log.Println("Błąd: Nie można uzyskać danych z kolekcji SRA", err)
		http.Error(w, "Błąd: Nie można uzyskać danych z kolekcji SRA", http.StatusInternalServerError)
		return
	}

	type Bus struct {
		Lp               int     `json:"lp"`
		Prefix           string  `json:"prefix"`
		StaticIdentifier *string `json:"static_identifier,omitempty"`
		Type             string  `json:"type"`
		Distance         *int    `json:"distance,omitempty"`
		ParkingMode      *string `json:"parking_mode,omitempty"`
	}
	type CongregationData struct {
		Name   string `json:"name"`
		Number int    `json:"number"`
		Lang   string `json:"lang"`
		Tura   int    `json:"tura"`
	}
	type ResponseData struct {
		Id           string           `json:"id"`
		Info         string           `json:"info"`
		Timestamp    string           `json:"timestamp"`
		Congregation CongregationData `json:"zbor"`
		Pilot1       db.Pilot         `json:"pilot1"`
		Pilot2       *db.Pilot        `json:"pilot2,omitempty"`
		Pilot3       *db.Pilot        `json:"pilot3,omitempty"`
	}
	var response []ResponseData
	for _, sra := range sraTableData {
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
			responseData := ResponseData{
				Id:        sra.ID.Hex(),
				Info:      "",
				Timestamp: sra.Timestamp.Time().Format("2006-01-02 15:04:05"),
				Congregation: CongregationData{
					Name:   congregation.Name,
					Number: congregation.Number,
					Lang:   congregation.Lang,
					Tura:   congregation.Tura,
				},
				Pilot1: pilots[0],
			}
			if len(pilots) > 1 {
				responseData.Pilot2 = &pilots[1]
			}
			if len(pilots) > 2 {
				responseData.Pilot3 = &pilots[2]
			}

			response = append(response, responseData)
		}
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Błąd: Nie można zakodować odpowiedzi JSON", err)
		return
	}
}
