package db

import (
	"gokongres/helpers"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// rozkład jazdy autokarów
type RJA struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SraID       primitive.ObjectID `bson:"sra_id" json:"sra_id"`
	SectorID    primitive.ObjectID `bson:"sector_id" json:"sid"`
	SectorOrder int                `bson:"sector_order" json:"sector_order"` // dawniej tura, teraz kolejność sektorów w rozkładzie jazdy
	A1          *string            `bson:"a1,omitempty" json:"a1,omitempty"`
	A2          *string            `bson:"a2,omitempty" json:"a2,omitempty"`
	A3          *string            `bson:"a3,omitempty" json:"a3,omitempty"`
	D1          *string            `bson:"d1,omitempty" json:"d1,omitempty"`
	D2          *string            `bson:"d2,omitempty" json:"d2,omitempty"`
	D3          *string            `bson:"d3,omitempty" json:"d3,omitempty"`
}

func (rja *RJA) WasArrived() bool {
	activeDay := helpers.GetActiveDay()
	switch activeDay {
	case time.Friday:
		return rja.A1 != nil
	case time.Saturday:
		return rja.A2 != nil
	case time.Sunday:
		return rja.A3 != nil
	}
	return false
}

func (rja *RJA) ArriveByDay(activeDay time.Weekday) string {
	switch activeDay {
	case time.Friday:
		if rja.A1 != nil {
			return *rja.A1
		}
	case time.Saturday:
		if rja.A2 != nil {
			return *rja.A2
		}
	case time.Sunday:
		if rja.A3 != nil {
			return *rja.A3
		}
	}
	return ""
}

func (rja *RJA) DepartureByDay(activeDay time.Weekday) string {
	switch activeDay {
	case time.Friday:
		if rja.D1 != nil {
			return *rja.D1
		}
	case time.Saturday:
		if rja.D2 != nil {
			return *rja.D2
		}
	case time.Sunday:
		if rja.D3 != nil {
			return *rja.D3
		}
	}
	return ""
}
