package monitoring

import (
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Get_TerminalsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collTerminals := db.Collection("terminals")
	if collTerminals == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'terminals' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'terminals' jest nil", http.StatusInternalServerError)
		return
	}

	opts := options.Find().SetProjection(bson.M{
		"_id": 1,
	})
	curTerm, err := collTerminals.Find(r.Context(), bson.M{}, opts)
	if err != nil {
		log.Printf("Błąd podczas pobierania terminali z bazy danych: %v", err)
		http.Error(w, "Błąd podczas pobierania terminali z bazy danych", http.StatusInternalServerError)
		return
	}
	defer curTerm.Close(r.Context())

	type RjaInfo struct {
		Id            primitive.ObjectID `json:"id"`
		BusIdentifier string             `json:"ident"`
		Name          string             `json:"name"`
		ZTura         string             `json:"ztura"`
		Lp            string             `json:"lp"`
		Tura          string             `json:"tura"`
		Arrive        string             `json:"arrive"`
		Departure     string             `json:"departure"`
	}
	type SectorInfo struct {
		Name string    `json:"name"`
		Rja  []RjaInfo `json:"rja,omitempty"`
	}
	type TerminalInfo struct {
		Name    string       `json:"name"`
		Sectors []SectorInfo `json:"sectors"`
	}
	var response []TerminalInfo

	var terminals []db.Terminal
	if err := curTerm.All(r.Context(), &terminals); err != nil {
		log.Printf("Błąd podczas dekodowania terminali z bazy danych: %v", err)
		http.Error(w, "Błąd podczas dekodowania terminali z bazy danych", http.StatusInternalServerError)
		return
	}

	rjaColl := db.Collection("rja")
	if rjaColl == nil {
		log.Println("Błąd połączenia z bazą danych: kolekcja 'rja' jest nil")
		http.Error(w, "Błąd połączenia z bazą danych: kolekcja 'rja' jest nil", http.StatusInternalServerError)
		return
	}

	for _, terminal := range terminals {
		resp := TerminalInfo{
			Name:    terminal.Name,
			Sectors: make([]SectorInfo, 0),
		}

		for _, sector := range terminal.Sectors {
			sinfo := SectorInfo{
				Name: sector.Name,
			}

			opts := options.Find().SetSort(bson.M{
				"tura": 1,
			})
			curRja, err := rjaColl.Find(r.Context(), bson.M{"sector_id": sector.Sid}, opts)
			if err != nil {
				log.Printf("Błąd podczas pobierania RJA z bazy danych: %v", err)
				http.Error(w, "Błąd podczas pobierania RJA z bazy danych", http.StatusInternalServerError)
				return
			}

			var rjaList []db.RJA
			if err := curRja.All(r.Context(), &rjaList); err != nil {
				log.Printf("Błąd podczas dekodowania RJA z bazy danych: %v", err)
				http.Error(w, "Błąd podczas dekodowania RJA z bazy danych", http.StatusInternalServerError)
				return
			}
			/*
				for _, rja := range rjaList {
					if !arriveToday(rja) {
						continue
					}

					sra := getSRAByID(rja.SraID)
					if sra == nil {
						log.Printf("Nie znaleziono SRA o ID: %v", rja.SraID)
						continue
					}

					congregation := getCongregationByID(sra.CongregationID)
					if congregation == nil {
						log.Printf("Nie znaleziono zbory o ID: %v", sra.CongregationID)
						continue
					}

					rjaInfo := RjaInfo{
						Id:            rja.ID,
						BusIdentifier: db.createShortBusID(&sra, sector.Name, rja.SectorOrder),
						Name:          congregation.Name,
					}

					sinfo.Rja = append(sinfo.Rja, rjaInfo)
				}
			*/
			resp.Sectors = append(resp.Sectors, sinfo)
		}

		response = append(response, resp)
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Błąd podczas kodowania odpowiedzi JSON: %v", err)
		http.Error(w, "Błąd podczas kodowania odpowiedzi JSON", http.StatusInternalServerError)
		return
	}
}
