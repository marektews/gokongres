package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUserByUsername pobiera użytkownika z bazy danych po nazwie użytkownika
func GetUserByUsername(ctx context.Context, username string) (*User, error) {
	coll := Collection("users")
	if coll == nil {
		err := fmt.Errorf("mongo client not initialized")
		log.Printf("Error in GetUserByUsername: %v", err)
		return nil, err
	}

	var user User
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		log.Printf("Error occurred while fetching user by username: %v", err)
		return nil, err
	}

	return &user, nil
}

// GetUserByID pobiera użytkownika z bazy danych po polu ID
func GetUserByID(ctx context.Context, uid primitive.ObjectID) (*User, error) {
	coll := Collection("users")
	if coll == nil {
		err := fmt.Errorf("mongo client not initialized")
		log.Printf("GetUserByID: error in GetUserByID: %v", err)
		return nil, err
	}

	var user User
	err := coll.FindOne(ctx, bson.M{"_id": uid}).Decode(&user)
	if err != nil {
		log.Printf("GetUserByID: error occurred while fetching user by ID: %v", err)
		return nil, err
	}

	return &user, nil
}

// GetUserFullName zwraca pełne imię i nazwisko użytkownika
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}
