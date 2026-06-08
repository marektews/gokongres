package sector

import (
	"encoding/json"
	"fmt"
	"gokongres/db"
	"log"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Schedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sectorID := r.PathValue("sector_id")

	collRJA := db.Collection("rja")
	if collRJA == nil {
		s := fmt.Sprintf("Collection 'rja' not found for sector_id '%s'", sectorID)
		log.Print(s)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}

	collCongregations := db.Collection("congegations")
	if collCongregations == nil {
		s := fmt.Sprintf("Collection 'congegations' not found for sector_id '%s'", sectorID)
		log.Print(s)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}

	collSectors := db.Collection("sectors")
	if collSectors == nil {
		s := fmt.Sprintf("Collection 'sectors' not found for sector_id '%s'", sectorID)
		log.Print(s)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}

	opts := options.Find().SetSort(bson.D{{Key: "sector_order", Value: 1}})
	cur, err := collRJA.Find(r.Context(), bson.M{"sector_id": sectorID}, opts)
	if err != nil {
		s := fmt.Sprintf("Error finding RJAs for sector_id '%s': %v", sectorID, err)
		log.Print(s)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}
	defer cur.Close(r.Context())

	var rjas []db.RJA
	if err := cur.All(r.Context(), &rjas); err != nil {
		s := fmt.Sprintf("Error decoding RJAs for sector_id '%s': %v", sectorID, err)
		log.Print(s)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}

	type BusInfo struct {
		Lp    *int   `json:"lp,omitempty"`
		Ident string `json:"ident"`
	}
	type CongregationInfo struct {
		Lang       string `json:"lang"`
		Name       string `json:"name"`
		Identifier string `json:"ident"`
	}
	type Times struct {
		Arrive    string `json:"arrive,omitempty"`
		Departure string `json:"departure,omitempty"`
	}
	type ScheduleInfo struct {
		SortOrder    int              `json:"tura"`
		D1           Times            `json:"d1,omitempty"`
		D2           Times            `json:"d2,omitempty"`
		D3           Times            `json:"d3,omitempty"`
		Bus          BusInfo          `json:"bus,omitempty"`
		Congregation CongregationInfo `json:"congregation,omitempty"`
	}

	sinfo := make([]ScheduleInfo, 0)
	for _, rja := range rjas {

		var sector db.Sector
		err = collSectors.FindOne(r.Context(), bson.M{"_id": rja.SectorID}).Decode(&sector)
		if err != nil {
			s := fmt.Sprintf("Error finding sector for sector_id '%s': %v", rja.SectorID, err)
			log.Print(s)
			http.Error(w, s, http.StatusInternalServerError)
			return
		}

		var sra db.SRA
		err = collSectors.FindOne(r.Context(), bson.M{"_id": rja.SraID}).Decode(&sra)
		if err != nil {
			s := fmt.Sprintf("Error finding SRA for sra_id '%s': %v", rja.SraID, err)
			log.Print(s)
			http.Error(w, s, http.StatusInternalServerError)
			return
		}

		var congregation db.Congregation
		err = collCongregations.FindOne(r.Context(), bson.M{"_id": sra.CongregationID}).Decode(&congregation)
		if err != nil {
			s := fmt.Sprintf("Error finding congregation for congregation_id '%s': %v", sra.CongregationID, err)
			log.Print(s)
			http.Error(w, s, http.StatusInternalServerError)
			return
		}

		si := ScheduleInfo{
			SortOrder: rja.SectorOrder,
			D1: Times{
				Arrive:    *rja.A1,
				Departure: *rja.D1,
			},
			D2: Times{
				Arrive:    *rja.A2,
				Departure: *rja.D2,
			},
			D3: Times{
				Arrive:    *rja.A3,
				Departure: *rja.D3,
			},
			Bus: BusInfo{
				Lp:    sra.Lp,
				Ident: strings.Replace(sector.Name, "x", fmt.Sprintf("%d", rja.SectorOrder), -1),
			},
			Congregation: CongregationInfo{
				Lang:       congregation.Lang,
				Name:       congregation.Name,
				Identifier: db.CreateShortBusID(&sra, sector.Name, rja.SectorOrder),
			},
		}
		sinfo = append(sinfo, si)
	}

	err = json.NewEncoder(w).Encode(sinfo)
	if err != nil {
		s := fmt.Sprintf("Error encoding schedule info for sector_id '%s': %v", sectorID, err)
		log.Print(s)
		http.Error(w, s, http.StatusInternalServerError)
		return
	}
}
