package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rozklad struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	SraID    int                `bson:"sra_id"`
	SektorID int                `bson:"sektor_id"`
	Tura     int                `bson:"tura"`
	A1       *time.Time         `bson:"a1,omitempty"`
	A2       *time.Time         `bson:"a2,omitempty"`
	A3       *time.Time         `bson:"a3,omitempty"`
	D1       *time.Time         `bson:"d1,omitempty"`
	D2       *time.Time         `bson:"d2,omitempty"`
	D3       *time.Time         `bson:"d3,omitempty"`
}
