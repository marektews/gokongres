package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// system rejestracji pojazdów
type SRP struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	CongregationID       primitive.ObjectID `bson:"congregation_id"`
	PassNr               int                `bson:"pass_nr"`
	Car1                 CarInfo            `bson:"car1"`
	Car2                 *CarInfo           `bson:"car2,omitempty"`
	Car3                 *CarInfo           `bson:"car3,omitempty"`
	MobilityRestrictions bool               `bson:"mobility_restrictions"`
	Timestamp            primitive.DateTime `bson:"timestamp"`
	D1                   *time.Time         `bson:"d1,omitempty"`
	D2                   *time.Time         `bson:"d2,omitempty"`
	D3                   *time.Time         `bson:"d3,omitempty"`
}

type CarInfo struct {
	RegNum string `bson:"regnum"`
	Lpg    bool   `bson:"lpg"`
}
