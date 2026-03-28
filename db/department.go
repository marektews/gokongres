package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Department struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Lang     string             `bson:"lang"`
	Name     string             `bson:"name"`
	Password string             `bson:"password"`
	Plimit   int                `bson:"plimit"`
	TuraID   primitive.ObjectID `bson:"tura"`
}
