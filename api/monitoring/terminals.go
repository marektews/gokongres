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

	activeTura := db.WhichTura(r.Context())
	if activeTura == nil {
		log.Println("Brak aktywnej tury")
		http.Error(w, "Brak aktywnej tury", http.StatusInternalServerError)
		return
	}

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
		Lp            *int               `json:"lp"`
	}
	type SectorInfo struct {
		Name string    `json:"name"`
		Rja  []RjaInfo `json:"rja"`
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
				Rja:  make([]RjaInfo, 0),
			}

			opts := options.Find().SetSort(bson.M{
				"sector_order": 1,
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
			for _, rja := range rjaList {
				var sra db.SRA
				if err := db.Collection("sra").FindOne(r.Context(), bson.M{"_id": rja.SraID}).Decode(&sra); err != nil {
					log.Printf("monitoring: nie znaleziono SRA dla rja %s: %v", rja.ID.Hex(), err)
					continue
				}

				// zbór musi należeć do aktywnej tury (tura == null oznacza dowolną)
				congFilter := bson.M{
					"$and": []bson.M{
						{"_id": sra.CongregationID},
						{"$or": []bson.M{{"tura": nil}, {"tura": activeTura.TID}}},
					},
				}
				var cong db.Congregation
				if err := db.Collection("congregations").FindOne(r.Context(), congFilter).Decode(&cong); err != nil {
					continue // zbór nie w aktywnej turze → pomijamy
				}

				sinfo.Rja = append(sinfo.Rja, RjaInfo{
					Id:            rja.ID,
					BusIdentifier: db.CreateShortBusID(&sra, sector.Name, rja.SectorOrder),
					Name:          cong.Name,
					Lp:            sra.Lp,
				})
			}

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
