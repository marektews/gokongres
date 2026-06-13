package sector

import (
	"encoding/json"
	"gokongres/db"
	"gokongres/helpers"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Initialize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sector_id := r.PathValue("sector_id")
	oid, err := primitive.ObjectIDFromHex(sector_id)
	if err != nil {
		log.Printf("Invalid sector ID '%s': %v", sector_id, err)
		http.Error(w, "Invalid sector ID", http.StatusBadRequest)
		return
	}

	activeTura := db.WhichTura(r.Context())
	if activeTura == nil {
		log.Println("No active tura found")
		http.Error(w, "No active tura found", http.StatusNotFound)
		return
	}

	// nazwa sektora (sektory są osadzone w dokumencie terminala)
	collTerm := db.Collection("terminals")
	if collTerm == nil {
		log.Println("Collection 'terminals' not found")
		http.Error(w, "Collection 'terminals' not found", http.StatusInternalServerError)
		return
	}

	var terminal db.Terminal
	err = collTerm.FindOne(r.Context(), bson.M{"sectors.sid": oid}).Decode(&terminal)
	if err != nil {
		log.Printf("Sector '%s' not found in any terminal: %v", sector_id, err)
		http.Error(w, "Sector not found", http.StatusNotFound)
		return
	}

	sectorName := ""
	for _, s := range terminal.Sectors {
		if s.Sid == oid {
			sectorName = s.Name
			break
		}
	}

	// autobusy przypisane do sektora (rozkład jazdy)
	collRJA := db.Collection("rja")
	if collRJA == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}

	collation := options.Collation{Locale: "pl", NumericOrdering: true, Strength: 1}
	rjaOpts := options.Find().SetSort(bson.D{{Key: "sector_order", Value: 1}}).SetCollation(&collation)
	cur, err := collRJA.Find(r.Context(), bson.M{"sector_id": oid}, rjaOpts)
	if err != nil {
		log.Printf("Error finding RJAs for sector_id '%s': %v", sector_id, err)
		http.Error(w, "Error finding RJAs", http.StatusInternalServerError)
		return
	}

	var allRJA []db.RJA
	if err := cur.All(r.Context(), &allRJA); err != nil {
		log.Printf("Error decoding RJAs for sector_id '%s': %v", sector_id, err)
		http.Error(w, "Error decoding RJAs", http.StatusInternalServerError)
		return
	}

	activeDay := helpers.GetActiveDay()
	collSRA := db.Collection("sra")
	collCong := db.Collection("congregations")

	type SraInfo struct {
		Lp *int `json:"lp"`
	}
	type CongregationInfo struct {
		Ident string `json:"ident"`
		Name  string `json:"name"`
	}
	type BusInfo struct {
		ID           primitive.ObjectID `json:"id"`
		Arrive       string             `json:"arrive"`
		Departure    string             `json:"departure"`
		Sra          SraInfo            `json:"sra"`
		Congregation CongregationInfo   `json:"congregation"`
	}
	type Response struct {
		SectorID primitive.ObjectID `json:"sid"`
		Name     string             `json:"name"`
		Buses    []BusInfo          `json:"buses"` // init [] → nigdy null
	}

	resp := Response{
		SectorID: oid,
		Name:     sectorName,
		Buses:    []BusInfo{},
	}

	for _, rja := range allRJA {
		if !rja.WasArrived() {
			continue
		}

		var sra db.SRA
		if err := collSRA.FindOne(r.Context(), bson.M{"_id": rja.SraID}).Decode(&sra); err != nil {
			log.Printf("Sector: SRA not found for rja %s: %v", rja.ID.Hex(), err)
			continue // pomijamy autobus bez SRA
		}

		// zbór musi być przypisany do aktywnej tury (tura == null oznacza dowolną)
		congregationFilter := bson.M{
			"$and": []bson.M{
				{"_id": sra.CongregationID},
				{"$or": []bson.M{{"tura": nil}, {"tura": activeTura.TID}}},
			},
		}
		var cong db.Congregation
		if err := collCong.FindOne(r.Context(), congregationFilter).Decode(&cong); err != nil {
			// zbór nie jest przypisany do tej tury — pomijamy ten autobus
			continue
		}

		resp.Buses = append(resp.Buses, BusInfo{
			ID:        rja.ID,
			Arrive:    rja.ArriveByDay(activeDay),
			Departure: rja.DepartureByDay(activeDay),
			Sra:       SraInfo{Lp: sra.Lp},
			Congregation: CongregationInfo{
				Ident: db.CreateShortBusID(&sra, sectorName, rja.SectorOrder),
				Name:  cong.Name,
			},
		})
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response for sector_id '%s': %v", sector_id, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
