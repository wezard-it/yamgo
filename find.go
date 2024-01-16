package yamgo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	P "github.com/gobeam/mongo-go-pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PopulateOptions struct {
	Collection string
	LocalField string
	As         string
	Projection []string
}

func (mf *Model) FindOne(filter bson.M, result interface{}) (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), MediumTimeout*time.Second)

	defer cancel()

	res := mf.col.FindOne(ctx, filter)

	if res.Err() != nil {
		return res.Err()
	}

	err = res.Decode(result)

	if err != nil {
		return err
	}

	return nil
}

func (mf *Model) FindByID(id string, result interface{}) (err error) {
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	return mf.FindOne(bson.M{"_id": objectID}, result)
}

func (mf *Model) FindByObjectID(objectID primitive.ObjectID, result interface{}) (err error) {

	if err != nil {
		return err
	}

	return mf.FindOne(bson.M{"_id": objectID}, result)
}

func (mf *Model) Find(filter bson.M, results interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), LongTimeout*time.Second)
	defer cancel()

	cur, err := mf.col.Find(ctx, filter)
	if err != nil {
		return err
	}

	if err = cur.All(ctx, results); err != nil {
		return err
	}

	return nil
}

func (mf *Model) executeCursorQuery(query []bson.M, sort bson.D, limit int64, collation *options.Collation, hint interface{}, projection string, lookups []PopulateOptions, results interface{}) error {

	options := options.Find()
	options.SetSort(sort)
	options.SetLimit(limit + 1)

	if collation != nil {
		options.SetCollation(collation)
	}

	if hint != nil {
		options.SetHint(hint)
	}

	if projection != "" {
		pMap := make(map[string]bool)
		str := strings.ReplaceAll(projection, "id", "_id")
		for _, key := range strings.Split(str, ",") {
			pMap[key] = true
		}
		options.SetProjection(pMap)
	}

	return mf.FindAndPopulate(bson.M{"$and": query}, *options, lookups, results)

	// return mf.FindWithOptions(bson.M{"$and": query}, *options, results)

}

func (mf *Model) PaginatedAggregate(example *[]bson.Raw, prevCursor string, nextCursor string, limit int64, pipeline ...interface{}) (Page, error) {

	ctx, cancel := context.WithTimeout(context.Background(), MediumTimeout*time.Second)

	defer cancel()

	var err error
	index := int64(1)

	var prev int64 = -1
	if prevCursor != "" {
		decodedCursor, err := decodeCursor(prevCursor)
		if err != nil {
			return Page{}, err
		}
		if len(decodedCursor) > 0 {
			value, ok := decodedCursor[0].Value.(int64)
			if !ok || value <= 0 {
				return Page{}, errors.New("invalid cursor")
			}
			prev = value
		}
	}

	var next int64 = -1
	if nextCursor != "" {
		decodedCursor, err := decodeCursor(nextCursor)
		if err != nil {
			return Page{}, err
		}
		if len(decodedCursor) > 0 {
			value, ok := decodedCursor[0].Value.(int64)
			if !ok || value <= 1 {
				return Page{}, errors.New("invalid cursor")
			}
			next = value
		}
	}

	if next > 1 {
		index = next
	} else if prev > 0 {
		index = prev
	}

	cur, err := P.New(mf.col).Context(ctx).Page(index).Limit(limit).Aggregate(pipeline...)
	if err != nil {
		return Page{}, err
	}
	index = cur.Pagination.Page

	oPrev := int64(-1)
	hasPrev := false
	oPrevCursor := ""
	if index > 1 {
		oPrev = index - 1
		hasPrev = true
		oPrevCursor, err = encodeCursor(bson.D{{Key: "page", Value: oPrev}})
		if err != nil {
			return Page{}, err
		}
	}

	var oNext int64 = -1
	hasNext := false
	oNextCursor := ""
	if index > 0 && index < cur.Pagination.Total {
		oNext = index + 1
		hasNext = true
		oNextCursor, err = encodeCursor(bson.D{{Key: "page", Value: oNext}})
		if err != nil {
			return Page{}, err
		}
	}

	*example = cur.Data

	page := Page{
		Previous:    oPrevCursor,
		HasPrevious: hasPrev,
		Next:        oNextCursor,
		HasNext:     hasNext,
		Count:       int(cur.Pagination.Total),
	}

	return page, nil
}

func (mf *Model) PaginatedFind(params PaginationFindParams, results interface{}) (Page, error) {

	var err error

	if results == nil {
		return Page{}, errors.New("results can't be nil")
	}

	params = ensureMandatoryParams(params)
	shouldSecondarySortOnID := params.PaginatedField != "_id"

	var count int
	if params.CountTotal {
		count, err = mf.CountDocuments(params.Query)
		if err != nil {
			return Page{}, err
		}
	}

	queries, sort, err := BuildQueries(params)

	if err != nil {
		return Page{}, err
	}

	err = mf.executeCursorQuery(queries, sort, params.Limit, params.Collation, params.Hint, params.Projection, params.Expansion, results)

	if err != nil {
		return Page{}, err
	}

	resultsPtr := reflect.ValueOf(results)
	resultsVal := resultsPtr.Elem()

	hasMore := resultsVal.Len() > int(params.Limit)

	if hasMore {
		resultsVal = resultsVal.Slice(0, resultsVal.Len()-1)
	}

	hasPrevious := params.Next != "" || (params.Previous != "" && hasMore)
	hasNext := params.Previous != "" || hasMore

	var previousCursor string
	var nextCursor string

	if resultsVal.Len() > 0 {
		if params.Previous != "" {
			for left, right := 0, resultsVal.Len()-1; left < right; left, right = left+1, right-1 {
				leftValue := resultsVal.Index(left).Interface()
				resultsVal.Index(left).Set(resultsVal.Index(right))
				resultsVal.Index(right).Set(reflect.ValueOf(leftValue))
			}
		}

		if hasPrevious {
			firstResult := resultsVal.Index(0).Interface()
			previousCursor, err = generateCursor(firstResult, params.PaginatedField, shouldSecondarySortOnID)
			if err != nil {
				return Page{}, fmt.Errorf("could not create a previous cursor: %s", err)
			}
		}

		if hasNext {
			lastResult := resultsVal.Index(resultsVal.Len() - 1).Interface()
			nextCursor, err = generateCursor(lastResult, params.PaginatedField, shouldSecondarySortOnID)
			if err != nil {
				return Page{}, fmt.Errorf("could not create a next cursor: %s", err)
			}
		}
	}

	page := Page{
		Previous:    previousCursor,
		HasPrevious: hasPrevious,
		Next:        nextCursor,
		HasNext:     hasNext,
		Count:       count,
	}

	resultsPtr.Elem().Set(resultsVal)

	return page, nil
}

func (mf *Model) FindWithOptions(filter bson.M, option options.FindOptions, results interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), LongTimeout*time.Second)

	defer cancel()

	cur, err := mf.col.Find(ctx, filter, &option)
	if err != nil {
		return err
	}
	err = cur.All(ctx, results)
	if err != nil {
		return err
	}
	return nil
}

func (mf *Model) FindOneAndPopulate(filter bson.M, findOptions options.FindOptions, populate []PopulateOptions, result interface{}) error {
	findOptions.SetLimit(-1)
	return mf.FindAndPopulate(filter, findOptions, populate, result)
}

func (mf *Model) FindAndPopulate(filter bson.M, option options.FindOptions, populate []PopulateOptions, results interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), LongTimeout*time.Second)

	defer cancel()

	var limit = 10

	pipeline := mongo.Pipeline{}

	if option.Limit != nil && *option.Limit > 0 {
		limit = int(*option.Limit)
	}

	matchStage := bson.D{
		{Key: "$match", Value: filter},
	}

	limitStage := bson.D{
		{Key: "$limit", Value: limit},
	}

	if option.Sort != nil {

		sortStage := bson.D{
			{Key: "$sort", Value: option.Sort},
		}
		pipeline = append(pipeline, sortStage)

	}

	pipeline = append(pipeline, matchStage, limitStage)

	if option.Projection != nil {
		projectionStage := bson.D{
			{Key: "$project", Value: option.Projection},
		}

		pipeline = append(pipeline, projectionStage)
	}

	for _, value := range populate {
		pipeline = append(pipeline, BuildLookupStage(value)...)
	}

	cur, err := mf.col.Aggregate(ctx, pipeline)

	if err != nil {
		return err
	}

	if err := cur.All(ctx, results); err != nil {
		return err
	}

	return nil
}

func (mf *Model) Aggregate(pipeline mongo.Pipeline, results interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), LongTimeout*time.Second)

	defer cancel()

	cur, err := mf.col.Aggregate(ctx, pipeline)

	if err != nil {
		return err
	}

	if err = cur.All(ctx, results); err != nil {
		return err
	}

	return nil
}

func BuildLookupStage(populate PopulateOptions) []bson.D {

	lookup := bson.D{
		{Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: populate.Collection},
				{Key: "localField", Value: populate.LocalField},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: populate.As},
			},
		},
	}

	addFields :=

		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: populate.As,
						Value: bson.D{
							{Key: "$cond",
								Value: bson.D{
									{Key: "if", Value: bson.D{{Key: "$isArray", Value: "$" + populate.LocalField}}},
									{Key: "then", Value: "$" + populate.As},
									{Key: "else", Value: bson.D{{Key: "$first", Value: "$" + populate.As}}},
								},
							},
						},
					},
				},
			},
		}

	expansion := []bson.D{lookup, addFields}

	return expansion
}
