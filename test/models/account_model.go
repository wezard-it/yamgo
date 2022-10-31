package models

import (
	"github.com/nocfer/yamgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountSchema struct {
	ID primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
}

func AccountModel() yamgo.Model {
	return yamgo.NewModel("accounts")
}
