package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
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
	log.Printf("Database URI: %s", uri)

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
		uri := os.Getenv("MONGODB_URI")
		if uri != "" {
			cs, err := connstring.ParseAndValidate(uri)
			if err == nil {
				log.Printf("Parse mongodb URI, database name: %s", cs.Database)
				dbname = cs.Database
			} else {
				log.Printf("Parse mongodb URI error: %v", err)
				dbname = "kongres"
			}
		} else {
			dbname = "kongres"
		}
	}
	log.Printf("Using MongoDB collection: %s.%s", dbname, name)
	return client.Database(dbname).Collection(name)
}

// Client zwraca aktualnego klienta MongoDB (może być nil).
func Client() *mongo.Client { return client }

// GetAllZbory pobiera wszystkie dokumenty z kolekcji "Zbory".
func GetAllZbory(ctx context.Context) ([]Congregation, error) {
	coll := Collection("congregations")
	if coll == nil {
		log.Print("GetAllZbory: mongo client not initialized")
		return nil, fmt.Errorf("mongo client not initialized")
	}

	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("GetAllZbory: Error finding zbory: %v", err)
		return nil, err
	}
	defer cur.Close(ctx)

	var results []Congregation
	for cur.Next(ctx) {
		var z Congregation
		if err := cur.Decode(&z); err != nil {
			log.Printf("GetAllZbory: Error decoding zbor: %v", err)
			return nil, err
		}
		results = append(results, z)
	}
	if err := cur.Err(); err != nil {
		log.Printf("GetAllZbory: Cursor error: %v", err)
		return nil, err
	}
	return results, nil
}

// UpdateZboryLimit aktualizuje pole plimit dokumentu Zbory o podanym id.
func UpdateZboryLimit(ctx context.Context, id string, plimit int) error {
	coll := Collection("congregations")
	if coll == nil {
		log.Print("UpdateZboryLimit: mongo client not initialized")
		return fmt.Errorf("mongo client not initialized")
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("UpdateZboryLimit: Error converting id: %v", err)
		return err
	}
	_, err = coll.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"plimit": plimit}})
	return err
}

// GetAllDepartments pobiera wszystkie dokumenty z kolekcji Department (Działy kongresowe).
func GetAllDepartments(ctx context.Context) ([]Department, error) {
	coll := Collection("departments")
	if coll == nil {
		log.Print("GetAllDepartments: mongo client not initialized")
		return nil, fmt.Errorf("mongo client not initialized")
	}

	cur, err := coll.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("GetAllDepartments: Error finding dzialy: %v", err)
		return nil, err
	}
	defer cur.Close(ctx)

	var results []Department
	for cur.Next(ctx) {
		var d Department
		if err := cur.Decode(&d); err != nil {
			log.Printf("GetAllDepartments: Error decoding dzial: %v", err)
			return nil, err
		}
		results = append(results, d)
	}
	if err := cur.Err(); err != nil {
		log.Printf("GetAllDepartments: Cursor error: %v", err)
		return nil, err
	}

	return results, nil
}

// UpdateDepartmentLimit aktualizuje plimit w kolekcji Department (Działy kongresowe).
func UpdateDepartmentLimit(ctx context.Context, id string, plimit int) error {
	coll := Collection("departments")
	if coll == nil {
		log.Print("UpdateDepartmentLimit: mongo client not initialized")
		return fmt.Errorf("mongo client not initialized")
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("UpdateDepartmentLimit: error converting id: %v", err)
		return err
	}

	_, err = coll.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"plimit": plimit}})
	log.Printf("UpdateDepartmentLimit: Updated department with id: %s, new plimit: %d", id, plimit)
	return err
}
