package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

type GenerateIdents struct {
	SRA bool `json:"sra"`
	SRP bool `json:"srp"`
	PK  bool `json:"pk"`
}
type ConstConfig struct {
	Tury           []Tura         `json:"tury"`
	GenerateIdents GenerateIdents `json:"generate_idents"`
}

/**
* Pobiera konfigurację stałą z bazy danych.
* Zwraca wskaźnik do ConstConfig i ewentualny błąd.
 */
func GetConstConfig() (*ConstConfig, error) {
	coll := Collection("const")
	if coll == nil {
		log.Print("GetAllTury: mongo client not initialized")
		return nil, fmt.Errorf("mongo client not initialized")
	}

	var config ConstConfig
	err := coll.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		log.Printf("GetConstConfig: error finding config info: %v", err)
		return nil, err
	}

	return &config, nil
}
