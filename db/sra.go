package db

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// system rejestracji autokarów
type SRA struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"-"`
	CongregationID primitive.ObjectID  `bson:"congregation_id" json:"congregation_id"`
	Bus            Bus                 `bson:"bus" json:"bus"`
	Lp             *int                `bson:"lp,omitempty" json:"lp"`
	Canceled       bool                `bson:"canceled" json:"canceled"`
	Pilot1ID       primitive.ObjectID  `bson:"pilot1_id" json:"pilot1_id"`
	Pilot2ID       *primitive.ObjectID `bson:"pilot2_id,omitempty" json:"pilot2_id"`
	Pilot3ID       *primitive.ObjectID `bson:"pilot3_id,omitempty" json:"pilot3_id"`
	Info           *string             `bson:"info,omitempty" json:"info"`
	Timestamp      primitive.DateTime  `bson:"timestamp" json:"timestamp"`

	// patrz moduł ADMSRA - pole stosowane zamiast Terminal.Prefix lub związanej z nim automatyki
	StaticIdentifier *string `bson:"static_identifier,omitempty" json:"static_identifier"` // identyfikator statyczny, np. "D1", "D2", "D3" - do przypisania ręcznie, nie jest generowany automatycznie
}

// wyjątek - sygnalizowanie, że zbór nie wynajmuje autokaru
type SRA_NoBus struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	CongregationID primitive.ObjectID `bson:"congregation_id" json:"congregation_id"`
	NoBus          bool               `bson:"nobus" json:"nobus"`
	Timestamp      primitive.DateTime `bson:"timestamp" json:"timestamp"`
}
