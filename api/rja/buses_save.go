package rja

import (
	"context"
	"encoding/json"
	"fmt"
	"gokongres/db"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RJAEntry struct {
	db.RJA
	Canceled bool `json:"canceled"`
}

func Get_BusesSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("rja.GetBusesSave called")

	type Request struct {
		SectorID string     `json:"sid"`
		TuraID   int        `json:"tura_id"`
		Buses    []RJAEntry `json:"rja"`
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding request for sector %s and tura %d: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	sectorID, err := primitive.ObjectIDFromHex(req.SectorID)
	if err != nil {
		log.Printf("Invalid sector ID %q: %v", req.SectorID, err)
		http.Error(w, "Invalid sector ID", http.StatusBadRequest)
		return
	}

	coll := db.Collection("rja")
	if coll == nil {
		log.Println("Collection 'rja' not found")
		http.Error(w, "Collection 'rja' not found", http.StatusInternalServerError)
		return
	}

	// Kasowanie poprzedniej listy autokarów dla tego sektora i tej tury.
	// Usuwamy cały zbiór wpisów, który był wyświetlany i edytowany - czyli
	// wpisy RJA tego sektora, których SRA należy do zboru przypisanego do
	// bieżącej tury. Dzięki temu wiersze usunięte w interfejsie znikają z
	// bazy i nie wracają po zapisie. (Wcześniej kasowane były tylko wiersze
	// nadal obecne w żądaniu, więc usunięty wiersz nigdy nie był kasowany.)
	sraIDs, err := getSRAIDsForTura(r.Context(), req.TuraID)
	if err != nil {
		log.Printf("Error resolving SRA IDs for tura %d: %v", req.TuraID, err)
		http.Error(w, "Error resolving SRA IDs for tura", http.StatusInternalServerError)
		return
	}
	_, err = coll.DeleteMany(r.Context(), bson.M{
		"sector_id": sectorID,
		"sra_id":    bson.M{"$in": sraIDs},
	})
	if err != nil {
		log.Printf("Error deleting old buses for sector %s and tura %d: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error deleting old buses for sector and tura", http.StatusInternalServerError)
		return
	}

	// wstawienie nowej listy autobusów
	buses := make([]any, len(req.Buses))
	for i, bus := range req.Buses {
		buses[i] = bus.RJA
		updateSRACanceledStatus(r.Context(), bus.SraID, bus.Canceled)
	}
	_, err = coll.InsertMany(r.Context(), buses)
	if err != nil {
		log.Printf("Error inserting new buses for sector %s and tura %d: %v", req.SectorID, req.TuraID, err)
		http.Error(w, "Error inserting new buses for sector and tura", http.StatusInternalServerError)
		return
	}
}

/**
 * Aktualizacja statusu anulowania SRA
 */
func updateSRACanceledStatus(ctx context.Context, sraID primitive.ObjectID, canceled bool) {
	coll := db.Collection("sra")
	if coll == nil {
		log.Println("Collection 'sra' not found")
		return
	}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": sraID}, bson.M{"$set": bson.M{"canceled": canceled}})
	if err != nil {
		log.Printf("Error updating SRA canceled status for ID %s: %v", sraID.Hex(), err)
	}
}

/**
 * Lista ID rekordów SRA należących do zborów przypisanych do danej tury
 * (tura == turaID lub tura == nil). Służy do ograniczenia kasowania rozkładu
 * jazdy do wpisów widocznych w bieżącej turze.
 */
func getSRAIDsForTura(ctx context.Context, turaID int) ([]primitive.ObjectID, error) {
	congregations, err := db.GetCongregationsForTura(ctx, strconv.Itoa(turaID))
	if err != nil {
		return nil, err
	}
	congIDs := db.GetCongregationIDs(congregations)

	coll := db.Collection("sra")
	if coll == nil {
		return nil, fmt.Errorf("collection 'sra' not found")
	}
	opts := options.Find().SetProjection(bson.M{"_id": 1})
	cur, err := coll.Find(ctx, bson.M{"congregation_id": bson.M{"$in": congIDs}}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var sras []db.SRA
	if err := cur.All(ctx, &sras); err != nil {
		return nil, err
	}
	ids := make([]primitive.ObjectID, len(sras))
	for i, sra := range sras {
		ids[i] = sra.ID
	}
	return ids, nil
}
