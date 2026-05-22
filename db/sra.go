package db

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// system rejestracji autokarów
type SRA struct {
	ID               primitive.ObjectID  `bson:"_id,omitempty"`
	CongregationID   primitive.ObjectID  `bson:"congregation_id"`
	Bus              Bus                 `bson:"bus"`
	Lp               *int                `bson:"lp,omitempty"`
	Canceled         bool                `bson:"canceled"`
	Pilot1ID         primitive.ObjectID  `bson:"pilot1_id"`
	Pilot2ID         *primitive.ObjectID `bson:"pilot2_id,omitempty"`
	Pilot3ID         *primitive.ObjectID `bson:"pilot3_id,omitempty"`
	Info             *string             `bson:"info,omitempty"`
	Timestamp        primitive.DateTime  `bson:"timestamp"`
	StaticIdentifier *string             `bson:"static_identifier,omitempty"` // identyfikator statyczny, np. "D1", "D2", "D3" - do przypisania ręcznie, nie jest generowany automatycznie
}

// wyjątek - sygnalizowanie, że zbór nie wynajmuje autokaru
type SRA_NoBus struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	CongregationID primitive.ObjectID `bson:"congregation_id"`
	NoBus          bool               `bson:"nobus"`
	Timestamp      primitive.DateTime `bson:"timestamp"`
}
