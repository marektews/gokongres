package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SOA struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	RjaID  int                `bson:"rja_id"`
	Status string             `bson:"status"`
	Ts     time.Time          `bson:"ts"`
}
