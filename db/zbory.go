package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Zbory struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Number int                `bson:"number"`
	Name   string             `bson:"name"`
	Lang   string             `bson:"lang"`
	Plimit int                `bson:"plimit"`
	Tura   int                `bson:"tura"`
}
