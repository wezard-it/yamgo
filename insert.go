package yamgo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func (mf *Model) InsertOne(record interface{}) (res *mongo.InsertOneResult, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), MediumTimeout*time.Second)

	defer cancel()
	res, err = mf.col.InsertOne(ctx, record)

	if err != nil {
		return nil, err
	}

	return res, err
}

func (mf *Model) InsertMany(records []interface{}) (res *mongo.InsertManyResult, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), LongTimeout*time.Second)
	defer cancel()

	res, err = mf.col.InsertMany(ctx, records)

	if err != nil {
		return nil, err
	}

	return res, err
}
