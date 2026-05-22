package db

import (
	"time"
)

type Tura struct {
	TID      int    `bson:"tid" json:"tid"`
	Shortcut string `bson:"shortcut" json:"shortcut"`
	Name     string `bson:"name" json:"name"`
	Range    struct {
		Begin time.Time `bson:"begin" json:"begin"`
		End   time.Time `bson:"end" json:"end"`
	} `bson:"range" json:"range"`
}
