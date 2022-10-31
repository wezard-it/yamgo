package test

import (
	"testing"

	"github.com/nocfer/yamgo/test/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCountDocuments(t *testing.T) {
	itemModel := models.ItemModel()

	items := []interface{}{models.ItemSchema{ID: primitive.NewObjectID()}, models.ItemSchema{ID: primitive.NewObjectID()}}

	_, err := itemModel.InsertMany(items)

	assert.Nil(t, err)

	result, err := itemModel.CountDocuments(bson.M{})

	assert.Nil(t, err)

	assert.Equal(t, result, 2)

	DropCollection("items")

}

func TestCountDocumentsWithFilter(t *testing.T) {
	itemModel := models.ItemModel()

	item := models.ItemSchema{ID: primitive.NewObjectID()}

	items := []interface{}{models.ItemSchema{ID: primitive.NewObjectID()}, item}

	_, err := itemModel.InsertMany(items)

	assert.Nil(t, err)

	result, err := itemModel.CountDocuments(bson.M{"_id": item.ID})

	assert.Nil(t, err)

	assert.Equal(t, result, 1)

	DropCollection("items")

}
