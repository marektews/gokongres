package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Terminale struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Name           string             `bson:"name"`
	IsBuffer       int                `bson:"is_buffer"`
	AssignedBuffer *int               `bson:"assigned_buffer,omitempty"`
}
