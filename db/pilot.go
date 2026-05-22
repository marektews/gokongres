package db

import "go.mongodb.org/mongo-driver/bson/primitive"

type Phone struct {
	CountryCode string `bson:"country_code" json:"country_code"`
	Number      string `bson:"number" json:"number"`
}

// dane pilota autokaru
type Pilot struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	FirstName string             `bson:"fn" json:"fn"`
	LastName  string             `bson:"ln" json:"ln"`
	Email     string             `bson:"email" json:"email"`
	Phone     Phone              `bson:"phone" json:"phone"`
}
