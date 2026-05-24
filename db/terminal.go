package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Terminal struct {
	ID      primitive.ObjectID `bson:"_id" json:"tid"`
	Name    string             `bson:"name" json:"name"`
	Sectors []Sector           `bson:"sectors,omitempty" json:"sectors,omitempty"`

	// prefiks na identyfikatorze autokaru (tura i sektor obliczane są automatycznie na podstawie rozkładu jazdy)
	// nadawany, gdy nie można wygenerować z 'Name' ponieważ są terminale z identyczną pierwszą literą
	Prefix *string `bson:"prefix,omitempty" json:"prefix,omitempty"`
}
