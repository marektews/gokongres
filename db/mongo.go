package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Connect nawiązuje połączenie z MongoDB. Jeśli uri jest puste, użyje zmiennej
// środowiskowej MONGODB_URI lub domyślnego mongodb://localhost:27017.
func Connect(ctx context.Context, uri string) error {
	if uri == "" {
		uri = os.Getenv("MONGODB_URI")
		if uri == "" {
			uri = "mongodb://localhost:27017"
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("mongo connect: %w", err)
	}

	if err := cli.Ping(ctx, nil); err != nil {
		return fmt.Errorf("mongo ping: %w", err)
	}

	client = cli
	return nil
}

// Disconnect zamyka połączenie z MongoDB.
func Disconnect(ctx context.Context) error {
	if client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return client.Disconnect(ctx)
}

// Collection zwraca wskaźnik do kolekcji w skonfigurowanej bazie.
// Nazwa bazy może być ustawiona przez zmienną MONGODB_DB, domyślnie "gokongres".
func Collection(name string) *mongo.Collection {
	if client == nil {
		return nil
	}
	dbname := os.Getenv("MONGODB_DB")
	if dbname == "" {
		dbname = "gokongres"
	}
	return client.Database(dbname).Collection(name)
}

// Client zwraca aktualnego klienta MongoDB (może być nil).
func Client() *mongo.Client { return client }

// InsertArrival wstawia dokument z polem message i zwraca id jako hex string oraz message.
func InsertArrival(ctx context.Context, message string) (map[string]interface{}, error) {
	coll := Collection("arrivals")
	if coll == nil {
		return nil, fmt.Errorf("mongo client not initialized")
	}
	res, err := coll.InsertOne(ctx, bson.M{"message": message})
	if err != nil {
		return nil, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("unexpected id type")
	}
	return map[string]interface{}{"id": oid.Hex(), "message": message}, nil
}

// GetAllArrivals pobiera wszystkie dokumenty z kolekcji arrivals i zwraca jako slice map.
func GetAllArrivals(ctx context.Context) ([]map[string]interface{}, error) {
	coll := Collection("arrivals")
	if coll == nil {
		return nil, fmt.Errorf("mongo client not initialized")
	}
	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []map[string]interface{}
	for cur.Next(ctx) {
		var d bson.M
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		for k, v := range d {
			if k == "_id" {
				if oid, ok := v.(primitive.ObjectID); ok {
					m["id"] = oid.Hex()
					continue
				}
			}
			m[k] = v
		}
		results = append(results, m)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
