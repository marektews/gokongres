package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Congregation struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Number int                `bson:"number" json:"number"`
	Name   string             `bson:"name" json:"name"`
	Lang   string             `bson:"lang" json:"lang"`
	Plimit int                `bson:"plimit" json:"plimit"`
	Tura   int                `bson:"tura" json:"tura"`
}
