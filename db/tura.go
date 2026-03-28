package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tura struct {
	ID       primitive.ObjectID `bson:"_id" json:"tura_id"`
	Shortcut string             `bson:"shortcut" json:"shortcut"`
	Name     string             `bson:"name" json:"name"`
	Range    struct {
		Begin time.Time `bson:"begin" json:"begin"`
		End   time.Time `bson:"end" json:"end"`
	} `bson:"range" json:"range"`
}
