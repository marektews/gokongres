package sra

import (
	"context"
	"encoding/json"
	"gokongres/db"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/**
*	Zapis zmodyfikowanych danych w module administracyjnym (ADMSRA).
*	Aktualizuje dane autokaru (typ, trasa, parking, lp, identyfikator statyczny),
*	dodatkowe informacje oraz dane pilotów (aktualizacja / dodanie / usunięcie).
 */
func Post_Save(w http.ResponseWriter, r *http.Request) {
	type RequestData struct {
		Id     string    `json:"id"`
		Info   string    `json:"info"`
		Bus    BusData   `json:"bus"`
		Pilot1 db.Pilot  `json:"pilot1"`
		Pilot2 *db.Pilot `json:"pilot2,omitempty"`
		Pilot3 *db.Pilot `json:"pilot3,omitempty"`
	}

	var req RequestData
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("sra.Save: decode request data error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sraID, err := primitive.ObjectIDFromHex(req.Id)
	if err != nil {
		log.Printf("sra.Save: invalid SRA ID '%s': %v", req.Id, err)
		http.Error(w, "Invalid SRA ID", http.StatusBadRequest)
		return
	}

	collSRA := db.Collection("sra")
	if collSRA == nil {
		log.Println("sra.Save: collection 'sra' not found")
		http.Error(w, "Collection 'sra' not found", http.StatusInternalServerError)
		return
	}
	collPilots := db.Collection("pilots")
	if collPilots == nil {
		log.Println("sra.Save: collection 'pilots' not found")
		http.Error(w, "Collection 'pilots' not found", http.StatusInternalServerError)
		return
	}

	// wczytanie istniejącego rekordu, aby poznać aktualne identyfikatory pilotów
	var sra db.SRA
	err = collSRA.FindOne(r.Context(), bson.M{"_id": sraID}).Decode(&sra)
	if err != nil {
		log.Printf("sra.Save: SRA '%s' not found: %v", req.Id, err)
		http.Error(w, "SRA not found", http.StatusNotFound)
		return
	}

	// pilot 1 istnieje zawsze - aktualizacja w miejscu
	if err := updatePilot(r.Context(), collPilots, sra.Pilot1ID, &req.Pilot1); err != nil {
		log.Printf("sra.Save: error updating pilot1: %v", err)
		http.Error(w, "Error updating pilot data", http.StatusInternalServerError)
		return
	}

	// piloci 2 i 3 są opcjonalni - aktualizacja / dodanie / usunięcie
	pilot2ID, err := resolvePilot(r.Context(), collPilots, sra.Pilot2ID, req.Pilot2)
	if err != nil {
		log.Printf("sra.Save: error resolving pilot2: %v", err)
		http.Error(w, "Error updating pilot data", http.StatusInternalServerError)
		return
	}
	pilot3ID, err := resolvePilot(r.Context(), collPilots, sra.Pilot3ID, req.Pilot3)
	if err != nil {
		log.Printf("sra.Save: error resolving pilot3: %v", err)
		http.Error(w, "Error updating pilot data", http.StatusInternalServerError)
		return
	}

	// budowanie aktualizacji dokumentu SRA
	setFields := bson.M{
		"bus.type":         req.Bus.Type,
		"bus.distance":     req.Bus.Distance,
		"bus.parking_mode": req.Bus.ParkingMode,
	}
	unsetFields := bson.M{}

	// lp: 0 / null oznacza brak numeracji
	if req.Bus.Lp != 0 {
		setFields["lp"] = req.Bus.Lp
	} else {
		unsetFields["lp"] = ""
	}

	// identyfikator statyczny: pusty oznacza powrót do automatyki
	if req.Bus.StaticIdentifier != "" {
		setFields["static_identifier"] = req.Bus.StaticIdentifier
	} else {
		unsetFields["static_identifier"] = ""
	}

	// dodatkowe informacje: puste czyszczą pole
	if req.Info != "" {
		setFields["info"] = req.Info
	} else {
		unsetFields["info"] = ""
	}

	if pilot2ID != nil {
		setFields["pilot2_id"] = *pilot2ID
	} else {
		unsetFields["pilot2_id"] = ""
	}
	if pilot3ID != nil {
		setFields["pilot3_id"] = *pilot3ID
	} else {
		unsetFields["pilot3_id"] = ""
	}

	update := bson.M{"$set": setFields}
	if len(unsetFields) > 0 {
		update["$unset"] = unsetFields
	}

	_, err = collSRA.UpdateOne(r.Context(), bson.M{"_id": sraID}, update)
	if err != nil {
		log.Printf("sra.Save: error updating SRA '%s': %v", req.Id, err)
		http.Error(w, "Error saving registration", http.StatusInternalServerError)
		return
	}

	log.Printf("sra.Save: successfully updated SRA '%s'", req.Id)
}

/**
*	Aktualizacja istniejącego dokumentu pilota.
 */
func updatePilot(ctx context.Context, coll *mongo.Collection, id primitive.ObjectID, p *db.Pilot) error {
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"fn":    p.FirstName,
		"ln":    p.LastName,
		"email": p.Email,
		"phone": p.Phone,
	}})
	return err
}

/**
*	Obsługa opcjonalnego pilota: aktualizacja istniejącego, dodanie nowego lub usunięcie.
*	Zwraca identyfikator pilota po operacji (nil, gdy pilota usunięto lub nie podano).
 */
func resolvePilot(ctx context.Context, coll *mongo.Collection, existingID *primitive.ObjectID, p *db.Pilot) (*primitive.ObjectID, error) {
	if p != nil {
		if existingID != nil {
			// aktualizacja istniejącego
			if err := updatePilot(ctx, coll, *existingID, p); err != nil {
				return nil, err
			}
			return existingID, nil
		}
		// dodanie nowego
		res, err := coll.InsertOne(ctx, db.Pilot{
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Email:     p.Email,
			Phone:     p.Phone,
		})
		if err != nil {
			return nil, err
		}
		newID := res.InsertedID.(primitive.ObjectID)
		return &newID, nil
	}

	// pilot nie został podany - usunięcie, jeśli istniał
	if existingID != nil {
		if _, err := coll.DeleteOne(ctx, bson.M{"_id": *existingID}); err != nil {
			return nil, err
		}
	}
	return nil, nil
}
