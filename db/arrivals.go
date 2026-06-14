package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Arrivals struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	BusID    primitive.ObjectID `bson:"bus_id"`
	DateTime time.Time          `bson:"datetime"`
	Arrived  bool               `bson:"arrived"`
}

// SetArrival upsertuje stan przyjazdu autokaru (po bus_id = id RJA).
func SetArrival(ctx context.Context, busID primitive.ObjectID, state bool) error {
	coll := Collection("arrivals")
	if coll == nil {
		return fmt.Errorf("collection 'arrivals' not found")
	}
	_, err := coll.UpdateOne(ctx,
		bson.M{"bus_id": busID},
		bson.M{"$set": bson.M{"arrived": state, "datetime": time.Now()}},
		options.Update().SetUpsert(true),
	)
	return err
}

// GetArrivalMap zwraca stany przyjazdów kluczowane po bus_id (hex).
func GetArrivalMap(ctx context.Context) (map[string]Arrivals, error) {
	res := make(map[string]Arrivals)
	coll := Collection("arrivals")
	if coll == nil {
		return res, nil
	}
	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var docs []Arrivals
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	for _, d := range docs {
		res[d.BusID.Hex()] = d
	}
	return res, nil
}
