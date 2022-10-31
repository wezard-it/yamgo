package test

import (
	"context"

	"github.com/nocfer/yamgo"
)

func DropCollection(c string) {

	if err := yamgo.GetCollection(c).Drop(context.TODO()); err != nil {
		panic(err)
	}
}

