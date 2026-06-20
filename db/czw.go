package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// czw - wydawanie zastępczych identyfikatorów parkingowych na czas wjazdu (moduł CZW)
type Czw struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	NrRej          string             `bson:"nr_rej" json:"nr_rej"`
	Phone          string             `bson:"phone" json:"phone"`
	NrIdent        int                `bson:"nr_ident" json:"nr_ident"`
	CongregationID primitive.ObjectID `bson:"congregation_id" json:"-"`
	Issuing        time.Time          `bson:"issuing" json:"issued"`
	Cancellation   *time.Time         `bson:"cancellation,omitempty" json:"cancellation"`
}
