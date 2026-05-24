package sra

type BusData struct {
	Lp               int    `json:"lp"`
	Prefix           string `json:"prefix"`
	StaticIdentifier string `json:"static_identifier"`
	Type             string `json:"type"`
	Distance         string `json:"distance"`
	ParkingMode      string `json:"parking_mode"`
}
