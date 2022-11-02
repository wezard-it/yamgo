package test

import (
	"testing"

	"github.com/nocfer/yamgo"
	"github.com/nocfer/yamgo/test/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestFindOneEmpty(t *testing.T) {

	itemModel := models.ItemModel()
	result := bson.M{}
	err := itemModel.FindOne(bson.M{}, &result)

	assert.Error(t, err)
	assert.Empty(t, result)

	DropCollection("items")
}

func TestFindOne(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}
	itemModel := models.ItemModel()

	_, err1 := itemModel.InsertOne(&item1)
	id2, err2 := itemModel.InsertOne(&item2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id2.InsertedID, item2.ID)
	result := models.ItemSchema{}

	err := itemModel.FindOne(bson.M{"_id": item2.ID}, &result)

	assert.Nil(t, err)

	assert.NotEmpty(t, result)
	assert.Equal(t, result.ID, id2.InsertedID)
	DropCollection("items")

}

func TestFindByID(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}
	itemModel := models.ItemModel()

	_, err1 := itemModel.InsertOne(&item1)
	id2, err2 := itemModel.InsertOne(&item2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id2.InsertedID, item2.ID)
	result := models.ItemSchema{}

	err := itemModel.FindByID(item2.ID.Hex(), &result)

	assert.Nil(t, err)

	assert.NotEmpty(t, result)
	assert.Equal(t, result.ID, id2.InsertedID)
	DropCollection("items")

}

func TestFindByObjectID(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}
	itemModel := models.ItemModel()

	_, err1 := itemModel.InsertOne(&item1)
	id2, err2 := itemModel.InsertOne(&item2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id2.InsertedID, item2.ID)
	result := models.ItemSchema{}

	err := itemModel.FindByObjectID(item2.ID, &result)

	assert.Nil(t, err)

	assert.NotEmpty(t, result)
	assert.Equal(t, result.ID, id2.InsertedID)

	DropCollection("items")

}

func TestFind(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}
	itemModel := models.ItemModel()

	id1, err1 := itemModel.InsertOne(&item1)
	id2, err2 := itemModel.InsertOne(&item2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, item1.ID)
	assert.Equal(t, id2.InsertedID, item2.ID)
	results := []models.ItemSchema{}

	err := itemModel.Find(bson.M{}, &results)

	assert.Nil(t, err)
	assert.NotEmpty(t, results)
	ids := []primitive.ObjectID{}

	for _, res := range results {
		ids = append(ids, res.ID)
	}

	assert.Contains(t, ids, id1.InsertedID)
	assert.Contains(t, ids, id2.InsertedID)

	DropCollection("items")

}

func TestPaginatedFind(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}
	itemModel := models.ItemModel()

	id1, err1 := itemModel.InsertOne(&item1)
	id2, err2 := itemModel.InsertOne(&item2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, item1.ID)
	assert.Equal(t, id2.InsertedID, item2.ID)
	results := []models.ItemSchema{}

	pfParams := yamgo.PaginationFindParams{
		Query:          bson.M{},
		Limit:          1,
		PaginatedField: "_id",
		CountTotal:     true,
		SortAscending:  true,
	}

	page, err := itemModel.PaginatedFind(pfParams, &results)
	assert.Nil(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, page.HasNext, true)
	assert.Equal(t, page.HasPrevious, false)
	assert.NotEmpty(t, page.Next)
	assert.Empty(t, page.Previous)
	assert.Equal(t, page.Count, 2)
	assert.Equal(t, results[0].ID, id1.InsertedID)

	DropCollection("items")

}

func TestFindWithOptions(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}
	itemModel := models.ItemModel()

	id1, err1 := itemModel.InsertOne(&item1)
	id2, err2 := itemModel.InsertOne(&item2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, item1.ID)
	assert.Equal(t, id2.InsertedID, item2.ID)
	results := []models.ItemSchema{}

	options := options.FindOptions{
		Sort: bson.M{"_id": -1},
	}

	err := itemModel.FindWithOptions(bson.M{}, options, &results)

	assert.Nil(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, results[1].ID, id1.InsertedID)
	assert.Equal(t, results[0].ID, id2.InsertedID)

	DropCollection("items")
}

func TestFindOneAndPopulate(t *testing.T) {
	item := models.ItemSchema{ID: primitive.NewObjectID()}
	foo := models.FooSchema{ID: primitive.NewObjectID(), Item: item.ID}

	itemModel := models.ItemModel()
	fooModel := models.FooModel()

	id1, err1 := itemModel.InsertOne(&item)
	id2, err2 := fooModel.InsertOne(&foo)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, item.ID)
	assert.Equal(t, id2.InsertedID, foo.ID)

	result := bson.M{}
	options := options.FindOptions{}
	populateOptions := []yamgo.PopulateOptions{{On: "items", Path: "item"}}

	err := fooModel.FindOneAndPopulate(bson.M{"_id": foo.ID}, options, populateOptions, &result)

	assert.Nil(t, err)
	assert.NotEmpty(t, result)

	assert.NotEmpty(t, result["item"])

	assert.Equal(t, result["_id"], foo.ID)
	assert.Equal(t, result["item"].(bson.M)["_id"], item.ID)

	DropCollection("items")
}

func TestFindAndPopulate(t *testing.T) {
	item := models.ItemSchema{ID: primitive.NewObjectID()}
	foo := models.FooSchema{ID: primitive.NewObjectID(), Item: item.ID}

	itemModel := models.ItemModel()
	fooModel := models.FooModel()

	id1, err1 := itemModel.InsertOne(&item)
	id2, err2 := fooModel.InsertOne(&foo)

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, item.ID)
	assert.Equal(t, id2.InsertedID, foo.ID)

	results := []bson.M{}
	options := options.FindOptions{}
	populateOptions := []yamgo.PopulateOptions{{On: "items", Path: "item"}}

	err := fooModel.FindAndPopulate(bson.M{"_id": foo.ID}, options, populateOptions, &results)

	assert.Nil(t, err)
	assert.NotEmpty(t, results)

	assert.NotEmpty(t, results[0]["item"])

	assert.Equal(t, results[0]["_id"], foo.ID)
	assert.Equal(t, results[0]["item"].(bson.M)["_id"], item.ID)

	DropCollection("items")
}

func TestAggregate(t *testing.T) {
	item1 := models.ItemSchema{ID: primitive.NewObjectID()}
	item2 := models.ItemSchema{ID: primitive.NewObjectID()}
	itemModel := models.ItemModel()

	id1, err1 := itemModel.InsertOne(&item1)
	id2, err2 := itemModel.InsertOne(&item2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, item1.ID)
	assert.Equal(t, id2.InsertedID, item2.ID)
	results := []models.ItemSchema{}

	pipeline := mongo.Pipeline{}

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: id1.InsertedID}}}}

	pipeline = append(pipeline, matchStage)

	err := itemModel.Aggregate(pipeline, &results)

	assert.Nil(t, err)
	assert.NotEmpty(t, results)
	assert.Len(t, results, 1)
	assert.Equal(t, results[0].ID, id1.InsertedID)

	DropCollection("items")

}
