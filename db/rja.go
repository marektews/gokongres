package db

import (
	"gokongres/helpers"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// rozkład jazdy autokarów
type RJA struct {
	ID          primitive.ObjectID `bson:"_id"`
	SraID       primitive.ObjectID `bson:"sra_id"`
	SectorID    primitive.ObjectID `bson:"sector_id"`
	SectorOrder int                `bson:"sector_order"` // dawniej tura, teraz kolejność sektorów w rozkładzie jazdy
	A1          *time.Time         `bson:"a1,omitempty"`
	A2          *time.Time         `bson:"a2,omitempty"`
	A3          *time.Time         `bson:"a3,omitempty"`
	D1          *time.Time         `bson:"d1,omitempty"`
	D2          *time.Time         `bson:"d2,omitempty"`
	D3          *time.Time         `bson:"d3,omitempty"`
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
			return rja.A1.Format("15:04")
		}
	case time.Saturday:
		if rja.A2 != nil {
			return rja.A2.Format("15:04")
		}
	case time.Sunday:
		if rja.A3 != nil {
			return rja.A3.Format("15:04")
		}
	}
	return ""
}

func (rja *RJA) DepartureByDay(activeDay time.Weekday) string {
	switch activeDay {
	case time.Friday:
		if rja.D1 != nil {
			return rja.D1.Format("15:04")
		}
	case time.Saturday:
		if rja.D2 != nil {
			return rja.D2.Format("15:04")
		}
	case time.Sunday:
		if rja.D3 != nil {
			return rja.D3.Format("15:04")
		}
	}
	return ""
}
