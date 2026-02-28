package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Dzialy struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Lang     string             `bson:"lang"`
	Name     string             `bson:"name"`
	Password string             `bson:"password"`
	Plimit   int                `bson:"plimit"`
	Tura     int                `bson:"tura"`
}
