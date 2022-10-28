package yamgo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	client   *mongo.Client
	Database *mongo.Database
	Err      error
}

type ConnectionParams struct {
	connectionUrl string
	dbName        string
}

const (
	ShortTimeout  time.Duration = 2
	MediumTimeout time.Duration = 5
	LongTimeout   time.Duration = 10
)

var _mongo Mongo

// It connects to the database.
func Connect(params ConnectionParams) {
	connectionURL := params.connectionUrl
	dbName := params.dbName

	if connectionURL == "" || dbName == "" {
		panic(errors.New("cannot start db, missing connection parameters"))
	}

	if _mongo.client == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_mongo.client, _mongo.Err = mongo.Connect(ctx, options.Client().ApplyURI(connectionURL))
		if _mongo.Err == nil {
			_mongo.Database = _mongo.client.Database(dbName)
			fmt.Printf("Successfully connected to db! (%s)\n", dbName)
		} else {
			panic(_mongo.Err)
		}
	}
}

func GetDB() Mongo {
	return _mongo
}

func GetCollection(collectionName string) *mongo.Collection {
	return _mongo.Database.Collection(collectionName)
}
