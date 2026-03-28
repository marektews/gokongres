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

// GetUserByUID pobiera użytkownika z bazy danych po polu Uid
func GetUserByUID(ctx context.Context, uid int) (*User, error) {
	coll := Collection("users")
	if coll == nil {
		err := fmt.Errorf("mongo client not initialized")
		log.Printf("Error in GetUserByUID: %v", err)
		return nil, err
	}

	var user User
	err := coll.FindOne(ctx, bson.M{"uid": uid}).Decode(&user)
	if err != nil {
		log.Printf("Error occurred while fetching user by UID: %v", err)
		return nil, err
	}

	return &user, nil
}

// GetUserByID pobiera użytkownika z bazy danych po ID
func GetUserByID(ctx context.Context, id string) (*User, error) {
	coll := Collection("users")
	if coll == nil {
		return nil, fmt.Errorf("mongo client not initialized")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	var user User
	err = coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserFullName zwraca pełne imię i nazwisko użytkownika
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}
