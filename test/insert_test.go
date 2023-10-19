package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wezard-it/yamgo/test/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestInsertOne(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}

	itemModel := models.ItemModel()
	id, err := itemModel.InsertOne(item1)

	assert.Nil(t, err)
	assert.Equal(t, item1.ID, id.InsertedID)

	DropCollection("items")
}

func TestInsertMany(t *testing.T) {

	itemModel := models.ItemModel()

	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}

	result, err := itemModel.InsertMany([]interface{}{item1, item2})

	assert.Nil(t, err)

	assert.Contains(t, result.InsertedIDs, item1.ID)
	assert.Contains(t, result.InsertedIDs, item2.ID)

	DropCollection("items")

}
