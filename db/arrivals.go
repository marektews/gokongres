package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Arrivals struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	BusID    int                `bson:"bus_id"`
	DateTime time.Time          `bson:"datetime"`
	Arrived  bool               `bson:"arrived"`
}
