package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sector struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Name   string             `bson:"name"`
	TuraID primitive.ObjectID `bson:"tura_id"`
}
