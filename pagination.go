// The MIT License

// Copyright (c) 2019-present QlikTech International AB

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package yamgo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Expansion struct {
		Collection   string
		LocalField   string
		ForeignField string
	}
	PaginationFindParams struct {
		Query          primitive.M        `form:"query"`
		Limit          int64              `form:"limit" binding:"required"`
		SortAscending  bool               `form:"sort_ascending"`
		PaginatedField string             `form:"paginated_field"`
		Collation      *options.Collation `form:"collation"`
		Next           string             `form:"next"`
		Previous       string             `form:"previous"`
		CountTotal     bool               `form:"count_total"`
		Hint           interface{}        `form:"hint"`
		Projection     string             `form:"projection"`
		Expansion      []PopulateOptions
	}

	Page struct {
		Previous    string `json:"previous,omitempty"`
		Next        string `json:"next,omitempty"`
		HasPrevious bool   `json:"has_previous"`
		HasNext     bool   `json:"has_next"`
		Count       int    `json:"count,omitempty"`
	}

	CursorError struct {
		err error
	}
)

func (e *CursorError) Error() string {
	return e.err.Error()
}

func GenerateCursorQuery(shouldSecondarySortOnID bool, paginatedField string, comparisonOp string, cursorFieldValues []interface{}) (map[string]interface{}, error) {

	var query map[string]interface{}
	if (shouldSecondarySortOnID && len(cursorFieldValues) != 2) ||
		(!shouldSecondarySortOnID && len(cursorFieldValues) != 1) {
		return nil, errors.New("wrong number of cursor field values specified")
	}

	if comparisonOp != "$lt" && comparisonOp != "$gt" {
		return nil, errors.New("invalid comparison operator specified: only $lt and $gt are allowed")
	}

	rangeOp := fmt.Sprintf("%se", comparisonOp)

	if shouldSecondarySortOnID {
		query = map[string]interface{}{"$or": []map[string]interface{}{
			{paginatedField: map[string]interface{}{comparisonOp: cursorFieldValues[0]}},
			{"$and": []map[string]interface{}{
				{paginatedField: map[string]interface{}{rangeOp: cursorFieldValues[0]}},
				{"_id": map[string]interface{}{comparisonOp: cursorFieldValues[1]}},
			}},
		}}
	} else {
		query = map[string]interface{}{paginatedField: map[string]interface{}{comparisonOp: cursorFieldValues[0]}}
	}
	return query, nil
}

func BuildQueries(p PaginationFindParams) (queries []bson.M, sort bson.D, err error) {
	p = ensureMandatoryParams(p)
	shouldSecondarySortOnID := p.PaginatedField != "_id"

	if p.Limit <= 0 {
		return []bson.M{}, nil, errors.New("a limit of at least 1 is required")
	}

	nextCursorValues, err := parseCursor(p.Next, shouldSecondarySortOnID)
	if err != nil {
		return []bson.M{}, nil, &CursorError{fmt.Errorf("next cursor parse failed: %s", err)}
	}

	previousCursorValues, err := parseCursor(p.Previous, shouldSecondarySortOnID)
	if err != nil {
		return []bson.M{}, nil, &CursorError{fmt.Errorf("previous cursor parse failed: %s", err)}
	}

	// Figure out the sort direction and comparison operator that will be used in the augmented query
	sortAsc := (!p.SortAscending && p.Previous != "") || (p.SortAscending && p.Previous == "")
	comparisonOp := "$gt"
	sortDir := 1
	if !sortAsc {
		comparisonOp = "$lt"
		sortDir = -1
	}

	queries = []bson.M{p.Query}

	// Setup the pagination query
	if p.Next != "" || p.Previous != "" {
		var cursorValues []interface{}
		if p.Next != "" {
			cursorValues = nextCursorValues
		} else if p.Previous != "" {
			cursorValues = previousCursorValues
		}
		var cursorQuery bson.M
		cursorQuery, err = GenerateCursorQuery(shouldSecondarySortOnID, p.PaginatedField, comparisonOp, cursorValues)
		if err != nil {
			return []bson.M{}, nil, err
		}
		queries = append(queries, cursorQuery)
	}

	// Setup the sort query
	if shouldSecondarySortOnID {
		sort = bson.D{{Key: p.PaginatedField, Value: sortDir}, {Key: "_id", Value: sortDir}}
	} else {
		sort = bson.D{{Key: "_id", Value: sortDir}}
	}

	return queries, sort, nil
}

func ensureMandatoryParams(p PaginationFindParams) PaginationFindParams {
	if p.PaginatedField == "" {
		p.PaginatedField = "_id"
		p.Collation = nil
	}

	return p
}

var parseCursor = func(cursor string, shouldSecondarySortOnID bool) ([]interface{}, error) {

	cursorValues := make([]interface{}, 0, 2)
	if cursor != "" {
		parsedCursor, err := decodeCursor(cursor)
		if err != nil {
			return nil, err
		}
		var id interface{}
		if shouldSecondarySortOnID {
			if len(parsedCursor) != 2 {
				return nil, errors.New("expecting a cursor with two elements")
			}
			paginatedFieldValue := parsedCursor[0].Value
			id = parsedCursor[1].Value
			cursorValues = append(cursorValues, paginatedFieldValue)
		} else {
			if len(parsedCursor) != 1 {
				return nil, errors.New("expecting a cursor with a single element")
			}
			id = parsedCursor[0].Value
		}
		cursorValues = append(cursorValues, id)
	}
	return cursorValues, nil
}

func decodeCursor(cursor string) (bson.D, error) {

	var cursorData bson.D
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return cursorData, err
	}

	err = bson.Unmarshal(data, &cursorData)
	return cursorData, err
}

func generateCursor(result interface{}, paginatedField string, shouldSecondarySortOnID bool) (string, error) {

	if result == nil {
		return "", fmt.Errorf("the specified result must be a non nil value")
	}
	// Handle pointer values and reduce number of times reflection is done on the same type.
	val := reflect.ValueOf(result)
	if val.Kind() == reflect.Ptr {
		_ = reflect.Indirect(val)
	}

	var recordAsBytes []byte
	var err error

	switch v := result.(type) {
	case []byte:
		recordAsBytes = v
	default:
		recordAsBytes, err = bson.Marshal(result)
		if err != nil {
			return "", err
		}
	}

	var recordAsMap map[string]interface{}
	err = bson.Unmarshal(recordAsBytes, &recordAsMap)
	if err != nil {
		return "", err
	}
	paginatedFieldValue := recordAsMap[paginatedField]
	// Set the cursor data
	cursorData := make(bson.D, 0, 2)
	cursorData = append(cursorData, bson.E{Key: paginatedField, Value: paginatedFieldValue})
	if shouldSecondarySortOnID {
		// Get the value of the ID field
		id := recordAsMap["_id"]
		cursorData = append(cursorData, bson.E{Key: "_id", Value: id})
	}
	// Encode the cursor data into a url safe string
	cursor, err := encodeCursor(cursorData)
	if err != nil {
		return "", fmt.Errorf("failed to encode cursor using %v: %s", cursorData, err)
	}
	return cursor, nil
}

func encodeCursor(cursorData bson.D) (string, error) {
	data, err := bson.Marshal(cursorData)
	return base64.RawURLEncoding.EncodeToString(data), err
}
