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

// GetAllUsers zwraca wszystkich użytkowników (kont admin).
func GetAllUsers(ctx context.Context) ([]User, error) {
	coll := Collection("users")
	if coll == nil {
		return nil, fmt.Errorf("collection 'users' not found")
	}
	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []User
	if err := cur.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// CreateUser wstawia nowego użytkownika (u.Hash ustawia wywołujący) i zwraca jego id.
func CreateUser(ctx context.Context, u *User) (primitive.ObjectID, error) {
	coll := Collection("users")
	if coll == nil {
		return primitive.NilObjectID, fmt.Errorf("collection 'users' not found")
	}
	res, err := coll.InsertOne(ctx, u)
	if err != nil {
		return primitive.NilObjectID, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("unexpected inserted id type")
	}
	return oid, nil
}

// UpdateUser ustawia (\$set) przekazane pola użytkownika o danym id.
func UpdateUser(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	coll := Collection("users")
	if coll == nil {
		return fmt.Errorf("collection 'users' not found")
	}
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

// DeleteUser usuwa użytkownika o danym id. Zwraca liczbę usuniętych dokumentów.
func DeleteUser(ctx context.Context, id primitive.ObjectID) (int64, error) {
	coll := Collection("users")
	if coll == nil {
		return 0, fmt.Errorf("collection 'users' not found")
	}
	res, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
