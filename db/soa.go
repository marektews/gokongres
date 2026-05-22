package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// system obsługi autokarów
type SOA struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	RjaID     primitive.ObjectID `bson:"rja_id"`
	Status    string             `bson:"status"`
	Timestamp time.Time          `bson:"ts"`
}
