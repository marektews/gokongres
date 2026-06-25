package db

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Congregation struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Number       int                `bson:"number" json:"number"`
	Name         string             `bson:"name" json:"name"`
	Lang         string             `bson:"lang" json:"lang"`
	Plimit       int                `bson:"plimit" json:"plimit"`
	Tura         int                `bson:"tura" json:"tura"`
	LimitRequest *LimitRequest      `bson:"limitRequest,omitempty" json:"limitRequest,omitempty"`
}

// LimitRequest to prośba zboru o zmianę limitu pojazdów (zapisywana z projektu srp).
type LimitRequest struct {
	Plimit int    `bson:"plimit" json:"plimit"`
	Reason string `bson:"reason" json:"reason"`
}

/**
 * Pobiera listę zborów przypisanych do danej tury
 */
func GetCongregationsForTura(ctx context.Context, turaId string) ([]Congregation, error) {
	coll := Collection("congregations")
	if coll == nil {
		log.Println("Collection 'congregations' not found")
		return nil, fmt.Errorf("collection 'congregations' not found")
	}
	turaID, err := strconv.Atoi(turaId)
	if err != nil {
		log.Printf("Error converting tura ID: %s to int, error: %v", turaId, err)
		return nil, err
	}
	cur, err := coll.Find(ctx, bson.M{"$or": []bson.M{{"tura": nil}, {"tura": turaID}}})
	if err != nil {
		log.Printf("Error finding congregations for tura ID: %s, error: %v", turaId, err)
		return nil, err
	}
	defer cur.Close(ctx)

	var congregations []Congregation
	err = cur.All(ctx, &congregations)
	if err != nil {
		log.Printf("Error decoding congregations for tura ID: %s, error: %v", turaId, err)
		return nil, err
	}
	return congregations, nil
}

/**
 * Tworzenie listy ID zborów z listy zborów
 */
func GetCongregationIDs(congregations []Congregation) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, len(congregations))
	for i, cong := range congregations {
		ids[i] = cong.ID
	}
	return ids
}
