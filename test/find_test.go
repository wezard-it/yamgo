package test

import (
	"testing"

	"github.com/nocfer/yamgo/test/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestFindOneEmpty(t *testing.T) {

	accountModel := models.AccountModel()
	result := bson.M{}
	err := accountModel.FindOne(bson.M{}, &result)

	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestFindOne(t *testing.T) {
	account1 := models.AccountSchema{ID: primitive.NewObjectID()}
	account2 := models.AccountSchema{ID: primitive.NewObjectID()}
	accountModel := models.AccountModel()

	_, err1 := accountModel.InsertOne(&account1)
	id2, err2 := accountModel.InsertOne(&account2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id2.InsertedID, account2.ID)
	result := models.AccountSchema{}

	err := accountModel.FindOne(bson.M{"_id": account2.ID}, &result)

	assert.Nil(t, err)

	assert.NotEmpty(t, result)
	assert.Equal(t, result.ID, id2.InsertedID)
}

func TestFindByID(t *testing.T) {
	account1 := models.AccountSchema{ID: primitive.NewObjectID()}
	account2 := models.AccountSchema{ID: primitive.NewObjectID()}
	accountModel := models.AccountModel()

	_, err1 := accountModel.InsertOne(&account1)
	id2, err2 := accountModel.InsertOne(&account2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id2.InsertedID, account2.ID)
	result := models.AccountSchema{}

	err := accountModel.FindByID(account2.ID.Hex(), &result)

	assert.Nil(t, err)

	assert.NotEmpty(t, result)
	assert.Equal(t, result.ID, id2.InsertedID)
}

func TestFindByObjectID(t *testing.T) {
	account1 := models.AccountSchema{ID: primitive.NewObjectID()}
	account2 := models.AccountSchema{ID: primitive.NewObjectID()}
	accountModel := models.AccountModel()

	_, err1 := accountModel.InsertOne(&account1)
	id2, err2 := accountModel.InsertOne(&account2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id2.InsertedID, account2.ID)
	result := models.AccountSchema{}

	err := accountModel.FindByObjectID(account2.ID, &result)

	assert.Nil(t, err)

	assert.NotEmpty(t, result)
	assert.Equal(t, result.ID, id2.InsertedID)
}

func TestFind(t *testing.T) {
	account1 := models.AccountSchema{ID: primitive.NewObjectID()}
	account2 := models.AccountSchema{ID: primitive.NewObjectID()}
	accountModel := models.AccountModel()

	id1, err1 := accountModel.InsertOne(&account1)
	id2, err2 := accountModel.InsertOne(&account2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, account1.ID)
	assert.Equal(t, id2.InsertedID, account2.ID)
	results := []models.AccountSchema{}

	err := accountModel.Find(bson.M{}, &results)

	assert.Nil(t, err)
	assert.NotEmpty(t, results)
	ids := []primitive.ObjectID{}

	for _, res := range results {
		ids = append(ids, res.ID)
	}

	assert.Contains(t, ids, id1.InsertedID)
	assert.Contains(t, ids, id2.InsertedID)
}

func TestPaginatedFind(t *testing.T) {
	// account1 := models.AccountSchema{ID: primitive.NewObjectID()}
	// account2 := models.AccountSchema{ID: primitive.NewObjectID()}
	// accountModel := models.AccountModel()

	// id1, err1 := accountModel.InsertOne(&account1)
	// id2, err2 := accountModel.InsertOne(&account2)
	// assert.Nil(t, err1)
	// assert.Nil(t, err2)
	// assert.Equal(t, id1.InsertedID, account1.ID)
	// assert.Equal(t, id2.InsertedID, account2.ID)
	// results := []models.AccountSchema{}

	// err := accountModel.Find(bson.M{}, &results)

	// assert.Nil(t, err)
	// assert.NotEmpty(t, results)
	// assert.Equal(t, results[0].ID, id1.InsertedID)
	// assert.Equal(t, results[1].ID, id2.InsertedID)
}

func TestFindWithOptions(t *testing.T) {
	account1 := models.AccountSchema{ID: primitive.NewObjectID()}
	account2 := models.AccountSchema{ID: primitive.NewObjectID()}
	accountModel := models.AccountModel()

	id1, err1 := accountModel.InsertOne(&account1)
	id2, err2 := accountModel.InsertOne(&account2)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, id1.InsertedID, account1.ID)
	assert.Equal(t, id2.InsertedID, account2.ID)
	results := []models.AccountSchema{}

	options := options.FindOptions{
		Sort: bson.M{"_id": -1},
	}

	err := accountModel.FindWithOptions(bson.M{}, options, &results)

	assert.Nil(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, results[1].ID, id1.InsertedID)
	assert.Equal(t, results[0].ID, id2.InsertedID)
}

func TestFindOneAndPopulate(t *testing.T) {

}

func TestFindAndPopulate(t *testing.T) {

}
