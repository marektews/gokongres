package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DepartmentPK struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	DepartmentID primitive.ObjectID `bson:"dzial_id"`
	PassNr       int                `bson:"pass_nr"`
	Regnum1      string             `bson:"regnum1"`
	Regnum2      *string            `bson:"regnum2,omitempty"`
	Regnum3      *string            `bson:"regnum3,omitempty"`
	Registered   time.Time          `bson:"registered"`
	D1           *time.Time         `bson:"d1,omitempty"`
	D2           *time.Time         `bson:"d2,omitempty"`
	D3           *time.Time         `bson:"d3,omitempty"`
}
