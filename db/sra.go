package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SRA struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	ZborID           int                `bson:"zbor_id"`
	BusID            int                `bson:"bus_id"`
	Lp               *int               `bson:"lp,omitempty"`
	Canceled         int                `bson:"canceled"`
	Pilot1ID         int                `bson:"pilot1_id"`
	Pilot2ID         *int               `bson:"pilot2_id,omitempty"`
	Pilot3ID         *int               `bson:"pilot3_id,omitempty"`
	Info             *string            `bson:"info,omitempty"`
	Timestamp        time.Time          `bson:"timestamp"`
	Prefix           string             `bson:"prefix"`
	StaticIdentifier *string            `bson:"static_identifier,omitempty"`
}
