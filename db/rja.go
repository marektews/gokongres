package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RJA struct {
	ID       primitive.ObjectID `bson:"_id"`
	SraID    primitive.ObjectID `bson:"sra_id"`
	SectorID primitive.ObjectID `bson:"sector_id"`
	TuraID   primitive.ObjectID `bson:"tura_id"`
	A1       *time.Time         `bson:"a1,omitempty"`
	A2       *time.Time         `bson:"a2,omitempty"`
	A3       *time.Time         `bson:"a3,omitempty"`
	D1       *time.Time         `bson:"d1,omitempty"`
	D2       *time.Time         `bson:"d2,omitempty"`
	D3       *time.Time         `bson:"d3,omitempty"`
}
