package models

import (
	"github.com/nocfer/yamgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FooSchema struct {
	ID   primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Item interface{}        `json:"item,omitempty" bson:"item,omitempty"`
}

func FooModel() yamgo.Model {
	return yamgo.NewModel("foos")
}
