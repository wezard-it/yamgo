package yamgo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Model struct {
	col *mongo.Collection
}

type Mongo struct {
	client   *mongo.Client
	Database *mongo.Database
	Err      error
}

type ConnectionParams struct {
	ConnectionUrl string
	DbName        string
}

const (
	ShortTimeout  time.Duration = 2
	MediumTimeout time.Duration = 5
	LongTimeout   time.Duration = 10
)

var _mongo Mongo

// It connects to the database.
func Connect(params ConnectionParams) {
	connectionURL := params.ConnectionUrl
	dbName := params.DbName

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

func Disconnect() error {
	fmt.Println("Disconnecting from DB")
	return _mongo.client.Disconnect(context.TODO())
}

func GetDB() Mongo {
	return _mongo
}

func GetCollection(collectionName string) *mongo.Collection {
	return _mongo.Database.Collection(collectionName)
}

func ToObjectID(hex string) primitive.ObjectID {
	oId, err := primitive.ObjectIDFromHex(hex)

	if err != nil {
		panic(err)
	}

	return oId
}

func NewModel(collectionName string) Model {
	return Model{col: GetCollection(collectionName)}
}
