package yamgo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type yamgo struct {
	col 	*mongo.Collection
}

func ToObjectID (hex string) primitive.ObjectID {
	oId, err := primitive.ObjectIDFromHex(hex)

	if err != nil {
		panic(err)
	}

	return oId
}

func NewModel(collectionName string) yamgo {
	return yamgo{col: GetCollection(collectionName)}
}
