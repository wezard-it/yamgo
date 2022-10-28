package yamgo

import (
	"context"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func (mf *Model) InsertOne(modelPtr interface{}) (res *mongo.InsertOneResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), MediumTimeout*time.Second)
	defer cancel()
	res, err = mf.col.InsertOne(ctx, modelPtr)
	if err != nil {
		return nil, err
	}
	// modelPtr.ID = res.InsertedID
	val := reflect.ValueOf(modelPtr).Elem().FieldByName("ID")
	val.Set(reflect.ValueOf(res.InsertedID))
	return res, err
}

//TODO Find a way to pass pointer and attach its ID to the respective array elements
func (mf *Model) InsertMany(models []interface{}) (res *mongo.InsertManyResult, err error) {
	// if models == nil || len(models) == 0 {
	// 	return nil, errors.New("The length of Model Array is 0")
	// }
	ctx, cancel := context.WithTimeout(context.Background(), LongTimeout*time.Second)
	defer cancel()
	// iM := make([]interface{}, 0)
	// iM = append(iM, models)
	res, err = mf.col.InsertMany(ctx, models)
	if err != nil {
		return nil, err
	}
	return res, err
}
