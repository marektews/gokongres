package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sektory struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
	Tid  int                `bson:"tid"`
}
