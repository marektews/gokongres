package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Terminal struct {
	ID      primitive.ObjectID `bson:"_id" json:"tid"`
	Name    string             `bson:"name" json:"name"`
	Sectors []Sector           `bson:"sectors,omitempty" json:"sectors,omitempty"`
}
