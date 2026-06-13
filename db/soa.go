package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// pojedyncza zmiana statusu autokaru
type SOAState struct {
	State string    `bson:"state" json:"state"`
	Ts    time.Time `bson:"ts" json:"ts"`
}

// system obsługi autokarów — jeden dokument na autokar (rja_id) z historią stanów
type SOA struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	RjaID  primitive.ObjectID `bson:"rja_id"`
	States []SOAState         `bson:"states"`
}

// Latest zwraca ostatni (najnowszy) stan oraz false, gdy lista jest pusta.
func (s *SOA) Latest() (SOAState, bool) {
	if len(s.States) == 0 {
		return SOAState{}, false
	}
	return s.States[len(s.States)-1], true
}

// PushSOAState dopisuje stan do dokumentu danego rja (upsert) i zwraca znacznik czasu.
func PushSOAState(ctx context.Context, rjaID primitive.ObjectID, state string) (time.Time, error) {
	coll := Collection("soa")
	if coll == nil {
		return time.Time{}, fmt.Errorf("collection 'soa' not found")
	}

	ts := time.Now()
	_, err := coll.UpdateOne(ctx,
		bson.M{"rja_id": rjaID},
		bson.M{"$push": bson.M{"states": SOAState{State: state, Ts: ts}}},
		options.Update().SetUpsert(true),
	)
	return ts, err
}
