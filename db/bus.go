package db

type Bus struct {
	Type        string `bson:"type" json:"type"`
	Distance    string `bson:"distance" json:"distance"`
	ParkingMode string `bson:"parking_mode" json:"parking_mode"`
}
