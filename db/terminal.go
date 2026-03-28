package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Terminal struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"terminal_id"`
	Name           string             `bson:"name" json:"name"`
	IsBuffer       int                `bson:"is_buffer" json:"is_buffer"`
	AssignedBuffer *int               `bson:"assigned_buffer,omitempty" json:"assigned_buffer,omitempty"`
}
