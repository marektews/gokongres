package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Pilot struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Fn    string             `bson:"fn"`
	Ln    string             `bson:"ln"`
	Email string             `bson:"email"`
	Phone string             `bson:"phone"`
}
