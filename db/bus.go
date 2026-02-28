package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Bus struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Type        string             `bson:"type"`
	Distance    string             `bson:"distance"`
	ParkingMode string             `bson:"parking_mode"`
}
