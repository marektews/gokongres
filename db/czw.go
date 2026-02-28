package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Czw struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	NrRej        string             `bson:"nr_rej"`
	Phone        string             `bson:"phone"`
	NrIdent      int                `bson:"nr_ident"`
	ZborID       int                `bson:"zbor_id"`
	Issuing      time.Time          `bson:"issuing"`
	Cancellation *time.Time         `bson:"cancellation,omitempty"`
}
