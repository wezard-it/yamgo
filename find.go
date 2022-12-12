package yamgo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PopulateOptions struct {
	Collection string
	LocalField string
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

	if *option.Limit != 0 {
		limit = int(*option.Limit)
	}

	matchStage := bson.D{
		{Key: "$match", Value: filter},
	}

	limitStage := bson.D{
		{Key: "$limit", Value: limit},
	}

	sortStage := bson.D{
		{Key: "$sort", Value: option.Sort},
	}

	pipeline := mongo.Pipeline{}
	pipeline = append(pipeline, sortStage, matchStage, limitStage)
	for _, value := range populate {
		pipeline = append(pipeline, buildLookupStage(value)...)
	}

	cur, err := mf.col.Aggregate(ctx, pipeline)

	if err != nil {
		return err
	}

	if option.Limit != nil && *option.Limit < 0 {
		if cur.Next(ctx) {

			if err := cur.Decode(results); err != nil {
				return err
			}
		}

	} else {
		err = cur.All(ctx, results)
	}

	if err != nil {
		return err
	}

	return nil
}

func (mf *Model) Aggregate(pipeline mongo.Pipeline, results interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), MediumTimeout*time.Second)

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

func buildLookupStage(populate PopulateOptions) []bson.D {

	var projection bson.D = nil

	if len(populate.Projection) > 0 {

		for _, projectionField := range populate.Projection {
			projection = append(projection, bson.E{Key: projectionField, Value: 1})
		}
	}

	lookupPipeline := bson.A{
		bson.D{
			{Key: "$match",
				Value: bson.D{
					{Key: "$expr",
						Value: bson.D{
							{Key: "$cond",
								Value: bson.D{
									{Key: "if", Value: bson.D{{Key: "$isArray", Value: "$$refId"}}},
									{Key: "then",
										Value: bson.D{
											{Key: "$in",
												Value: bson.A{
													"$_id",
													"$$refId",
												},
											},
										},
									},
									{Key: "else",
										Value: bson.D{
											{Key: "$eq",
												Value: bson.A{
													"$_id",
													"$$refId",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.D{
			{Key: "$project",
				Value: bson.D{
					{Key: "__v", Value: 0},
					{Key: "v", Value: 0},
					{Key: "updated_at", Value: 0},
					{Key: "created_at", Value: 0},
				},
			},
		},
	}

	if projection != nil {
		lookupPipeline = append(lookupPipeline, bson.D{
			{Key: "$project", Value: projection},
		})

	}

	lookup := bson.D{
		{Key: "$lookup",
			Value: bson.D{
				{Key: "from", Value: populate.Collection},
				{Key: "let", Value: bson.D{{Key: "refId", Value: "$" + populate.LocalField}}},
				{Key: "pipeline",
					Value: lookupPipeline,
				},
				{Key: "as", Value: "tmp_" + populate.LocalField},
			},
		},
	}

	addFields :=

		bson.D{
			{Key: "$addFields",
				Value: bson.D{
					{Key: populate.LocalField,
						Value: bson.D{
							{Key: "$cond",
								Value: bson.D{
									{Key: "if", Value: bson.D{{Key: "$isArray", Value: "$" + populate.LocalField}}},
									{Key: "then", Value: "$tmp_" + populate.LocalField},
									{Key: "else", Value: bson.D{{Key: "$first", Value: "$tmp_" + populate.LocalField}}},
								},
							},
						},
					},
				},
			},
		}

	unset := bson.D{
		{Key: "$unset",
			Value: bson.A{
				"tmp_" + populate.LocalField,
			},
		},
	}

	expansion := []bson.D{lookup, addFields, unset}

	return expansion
}
