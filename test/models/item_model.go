package models

import (
	"github.com/wezard-it/yamgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemSchema struct {
	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
}

func ItemModel() yamgo.Model {
	return yamgo.NewModel("items")
}
