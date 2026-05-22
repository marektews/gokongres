package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sector struct {
	Sid  primitive.ObjectID `bson:"sid" json:"sid"`
	Name string             `bson:"name" json:"name"`
}
