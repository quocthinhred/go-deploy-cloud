package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gitlab.com/thuocsi.vn-sdk/go-sdk/sdk/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Instance struct {
	ColName        string
	DBName         string
	TemplateObject interface{}

	db  *mongo.Database
	col *mongo.Collection
}

// convertToObject convert bson to object
func (m *Instance) ApplyDatabase(database *mongo.Database) *Instance {
	m.db = database
	m.col = database.Collection(m.ColName)
	m.DBName = database.Name()
	return m
}

// convertToObject convert bson to object
func (m *Instance) convertToObject(b bson.M) (interface{}, error) {
	obj := m.newObject()

	if b == nil {
		return obj, nil
	}

	bytes, err := bson.Marshal(b)
	if err != nil {
		return nil, err
	}

	bson.Unmarshal(bytes, obj)
	return obj, nil
}

// convertToBson Go object to map (to get / query)
func (m *Instance) convertToBson(ent interface{}) (bson.M, error) {
	if ent == nil {
		return bson.M{}, nil
	}

	sel, err := bson.Marshal(ent)
	if err != nil {
		return nil, err
	}

	obj := bson.M{}
	bson.Unmarshal(sel, &obj)

	return obj, nil
}

// newObject return new object with same type of TemplateObject
func (m *Instance) newObject() interface{} {
	t := reflect.TypeOf(m.TemplateObject)
	// fmt.Println(t)
	v := reflect.New(t)
	return v.Interface()
}

// newList return new object with same type of TemplateObject
func (m *Instance) newList(limit int) interface{} {
	t := reflect.TypeOf(m.TemplateObject)
	return reflect.MakeSlice(reflect.SliceOf(t), 0, limit).Interface()
}

func (m *Instance) interfaceSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil, errors.New("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret, nil
}

func (m *Instance) parseSingleResult(result *mongo.SingleResult, action string) *common.APIResponse {
	// parse result
	obj := m.newObject()
	err := result.Decode(obj)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "MAP_OBJECT_FAILED",
		}
	}

	// put to slice
	list := m.newList(1)
	listValue := reflect.Append(reflect.ValueOf(list),
		reflect.Indirect(reflect.ValueOf(obj)))

	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: action + " " + m.ColName + " successfully.",
		Data:    listValue.Interface(),
	}
}

// Create insert one object into DB
func (m *Instance) Create(entity interface{}) *common.APIResponse {

	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert to bson
	obj, err := m.convertToBson(entity)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "MAP_OBJECT_FAILED",
		}
	}

	// init time
	if obj["created_time"] == nil {
		obj["created_time"] = time.Now()
	}

	// insert
	result, err := m.col.InsertOne(context.TODO(), obj)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB Error: " + err.Error(),
		}
	}

	obj["_id"] = result.InsertedID
	entity, _ = m.convertToObject(obj)

	list := m.newList(1)
	listValue := reflect.Append(reflect.ValueOf(list),
		reflect.Indirect(reflect.ValueOf(entity)))

	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Create " + m.ColName + " successfully.",
		Data:    listValue.Interface(),
	}

}

// CreateMany insert many object into db
func (m *Instance) CreateMany(entityList interface{}) *common.APIResponse {

	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Create many - Collection " + m.ColName + " is not init.",
		}
	}

	list, err := m.interfaceSlice(entityList)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Create many - Invalid slice.",
		}
	}

	var bsonList []interface{}
	now := time.Now()
	for _, item := range list {
		b, err := m.convertToBson(item)
		if err != nil {
			return &common.APIResponse{
				Status:  common.APIStatus.Error,
				Message: "DB error: Create many - Invalid bson object.",
			}
		}
		if b["created_time"] == nil {
			b["created_time"] = now
		}
		bsonList = append(bsonList, b)
	}

	result, err := m.col.InsertMany(context.TODO(), bsonList)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "CREATE_FAILED",
		}
	}

	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Create " + m.ColName + "(s) successfully.",
		Data:    result.InsertedIDs,
	}
}

// Query Get all object in DB
func (m *Instance) Query(query interface{}, offset int64, limit int64, sortFields *bson.M) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}
	opt := &options.FindOptions{}
	k := int64(1000)
	if limit <= 0 {
		opt.Limit = &k
	} else {
		opt.Limit = &limit
	}
	if offset > 0 {
		opt.Skip = &offset
	}
	if sortFields != nil {
		opt.Sort = sortFields
	}

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: QueryOne - Cannot convert object - " + err.Error(),
		}
	}

	result, err := m.col.Find(context.TODO(), converted, opt)

	if err != nil || result.Err() != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ".",
			ErrorCode: "NOT_FOUND",
		}
	}

	list := m.newList(int(limit))
	err = result.All(context.TODO(), &list)
	result.Close(context.TODO())
	if err != nil || reflect.ValueOf(list).Len() == 0 {
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ".",
			ErrorCode: "NOT_FOUND",
		}
	}

	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Query " + m.ColName + " successfully.",
		Data:    list,
	}
}

// Query Get all object in DB
func (m *Instance) QueryAll() *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB error: Collection " + m.ColName + " is not init.",
			ErrorCode: "NOT_INIT_YET",
		}
	}
	rs, err := m.col.Find(context.TODO(), bson.M{})
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any " + m.ColName + ".",
			ErrorCode: "NOT_FOUND",
		}
	}

	list := m.newList(1000)
	rs.All(context.TODO(), &list)
	rs.Close(context.TODO())
	if reflect.ValueOf(list).Len() == 0 {
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ".",
			ErrorCode: "NOT_FOUND",
		}
	}
	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Query " + m.ColName + " successfully.",
		Data:    list,
	}
}

// QueryOne ...
func (m *Instance) QueryOne(query interface{}) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: QueryOne - Cannot convert object - " + err.Error(),
		}
	}

	// do find
	result := m.col.FindOne(context.TODO(), converted)

	if result == nil || result.Err() != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ".",
			ErrorCode: "NOT_FOUND",
		}
	}

	return m.parseSingleResult(result, "Query")
}

// Update Update all matched item
func (m *Instance) UpdateMany(query interface{}, updater interface{}) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	obj, err := m.convertToBson(updater)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "MAP_OBJECT_FAILED",
		}
	}
	obj["last_updated_time"] = time.Now()

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: QueryOne - Cannot convert object - " + err.Error(),
		}
	}

	// do update
	result, err := m.col.UpdateMany(context.TODO(), converted, bson.M{
		"$set": obj,
	})
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "Update error: UpdateMany - " + err.Error(),
			ErrorCode: "UPDATE_FAILED",
		}
	}

	if result.MatchedCount == 0 {
		return &common.APIResponse{
			Status:  common.APIStatus.Ok,
			Message: "Not found any " + m.ColName + ".",
		}
	}

	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Update " + m.ColName + " successfully.",
	}
}

// UpdateOne Update one matched object.
func (m *Instance) UpdateOne(query interface{}, updater interface{}, opts ...*options.FindOneAndUpdateOptions) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert
	bUpdater, err := m.convertToBson(updater)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "MAP_OBJECT_FAILED",
		}
	}
	bUpdater["last_updated_time"] = time.Now()

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: UpdateOne - Cannot convert object - " + err.Error(),
		}
	}

	// do update
	if opts == nil {
		after := options.After
		opts = []*options.FindOneAndUpdateOptions{
			{
				ReturnDocument: &after,
			},
		}
	}
	result := m.col.FindOneAndUpdate(context.TODO(), converted, bson.M{"$set": bUpdater}, opts...)
	if result.Err() != nil {
		detail := ""
		if result != nil {
			detail = result.Err().Error()
		}
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ". Error detail: " + detail,
			ErrorCode: "NOT_FOUND",
		}
	}

	return m.parseSingleResult(result, "UpdateOne")
}

// UpdateOne Update one matched object.
func (m *Instance) Upsert(query interface{}, updater interface{}) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert
	bUpdater, err := m.convertToBson(updater)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "MAP_OBJECT_FAILED",
		}
	}
	bUpdater["last_updated_time"] = time.Now()
	createdTime, ok := bUpdater["created_time"]
	if !ok || createdTime == nil {
		createdTime = bUpdater["last_updated_time"]
	} else {
		delete(bUpdater, "created_time")
	}

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: UpdateOne - Cannot convert object - " + err.Error(),
		}
	}

	// do update
	after := options.After
	t := true
	opts := []*options.FindOneAndUpdateOptions{
		{
			ReturnDocument: &after,
			Upsert:         &t,
		},
	}

	if bUpdater["_id"] != nil {
		delete(bUpdater, "_id")
	}
	result := m.col.FindOneAndUpdate(context.TODO(), converted, bson.M{
		"$set": bUpdater,
		"$setOnInsert": bson.M{
			"created_time": createdTime,
		},
	}, opts...)
	if result.Err() != nil {
		detail := ""
		if result != nil {
			detail = result.Err().Error()
		}
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ". Error detail: " + detail,
			ErrorCode: "NOT_FOUND",
		}
	}

	return m.parseSingleResult(result, "UpdateOne")
}

// UpdateOneWithOption Update one with option $set, $inc ... matched object.
func (m *Instance) UpdateOneWithOption(query interface{}, updater interface{}, opts ...*options.FindOneAndUpdateOptions) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert
	bUpdater, err := m.convertToBson(updater)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "MAP_OBJECT_FAILED",
		}
	}

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: UpdateOne - Cannot convert object - " + err.Error(),
		}
	}

	// do update
	result := m.col.FindOneAndUpdate(context.TODO(), converted, bUpdater, opts...)
	if result.Err() != nil {
		detail := ""
		if result != nil {
			detail = result.Err().Error()
		}
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ". Error detail: " + detail,
			ErrorCode: "NOT_FOUND",
		}
	}

	return m.parseSingleResult(result, "UpdateOne")
}

// ReplaceOneWithOption Replace one matched object with option.
func (m *Instance) ReplaceOneWithOption(query interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert
	bReplacement, err := m.convertToBson(replacement)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "DB Error: " + err.Error(),
			ErrorCode: "MAP_OBJECT_FAILED",
		}
	}

	if bReplacement["created_time"] == nil {
		bReplacement["created_time"] = time.Now()
	}
	bReplacement["last_updated_time"] = time.Now()

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: ReplaceOne - Cannot convert object - " + err.Error(),
		}
	}

	// do replace
	result := m.col.FindOneAndReplace(context.TODO(), converted, bReplacement, opts...)
	if result.Err() != nil {
		detail := ""
		if result != nil {
			detail = result.Err().Error()
		}
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ". Error detail: " + detail,
			ErrorCode: "NOT_FOUND",
		}
	}

	return m.parseSingleResult(result, "ReplaceOne")
}

// Delete Delete all object which matched with selector
func (m *Instance) Delete(query interface{}) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert query
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Delete - Cannot convert object - " + err.Error(),
		}
	}

	_, err = m.col.DeleteMany(context.TODO(), converted)
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "Delete error: " + err.Error(),
			ErrorCode: "DELETE_FAILED",
		}
	}
	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Delete " + m.ColName + " successfully.",
	}
}

// Count Count object which matched with query.
func (m *Instance) Count(query interface{}) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert query
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Count - Cannot convert object - " + err.Error(),
		}
	}

	// if query is empty -> count by EstimatedDocumentCount else count by CountDocuments
	count := int64(0)
	if len(converted) == 0 {
		count, err = m.col.EstimatedDocumentCount(context.TODO(), nil)
	} else {
		count, err = m.col.CountDocuments(context.TODO(), converted)
	}
	if err != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.Error,
			Message:   "Count error: " + err.Error(),
			ErrorCode: "COUNT_FAILED",
		}
	}

	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Count query executed successfully.",
		Total:   count,
	}

}

// IncreOne Increase one field of the document & return new value
func (m *Instance) IncreOne(query interface{}, fieldName string, value int) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	t := true
	after := options.After
	updater := bson.M{
		"$inc": bson.D{
			{fieldName, value},
		},
	}
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &t,
	}

	// convert query
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: IncreOne - Cannot convert object - " + err.Error(),
		}
	}

	result := m.col.FindOneAndUpdate(context.TODO(), converted, updater, &opt)

	return m.parseSingleResult(result, "Incre "+fieldName+" of")
}

// CreateIndex ...
func (m *Instance) CreateIndex(keys bson.D, options *options.IndexOptions) error {
	_, err := m.col.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    keys,
		Options: options,
	})
	return err
}

// Aggregate ...
func (m *Instance) Aggregate(pipeline interface{}, result interface{}) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	q, err := m.col.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Aggregate - " + err.Error(),
		}
	}
	err = q.All(context.TODO(), result)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Aggregate - " + err.Error(),
		}
	}

	return &common.APIResponse{
		Status: common.APIStatus.Ok,
	}
}

// Aggregate ...
func (m *Instance) Distinct(filter interface{}, field string, opt ...*options.DistinctOptions) *common.APIResponse {
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// convert query
	converted, err := m.convertToBson(filter)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Distinct - Cannot convert object - " + err.Error(),
		}
	}

	result, err := m.col.Distinct(context.TODO(), field, converted, opt...)

	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Distinct " + err.Error(),
		}
	}
	return &common.APIResponse{
		Status: common.APIStatus.Ok,
		Data:   result,
	}
}

// Get Client
func (m *Instance) GetClient() *mongo.Client {
	return m.db.Client()
}

// GetChangeStream func
func (m *Instance) GetChangeStream(dbName string, collectionName string, cb func(bson.M)) (err error) {

	opts := options.ChangeStream()
	opts.SetFullDocument(options.UpdateLookup)

	return m.GetChangeStreamWithOpt(dbName, collectionName, opts, cb)
}

// GetChangeStreamWithOpt Get opslog with change stream option
func (m *Instance) GetChangeStreamWithOpt(dbName string, collectionName string, opts *options.ChangeStreamOptions, cb func(bson.M)) (err error) {

	ctx := context.Background()
	cur := &mongo.ChangeStream{}
	pipelineData := []bson.D{}

	client := m.db.Client()

	if dbName != "" && collectionName != "" { // Watching a collection
		fmt.Println("Watching", dbName+"."+collectionName)

		coll := client.Database(dbName).Collection(collectionName)
		cur, err = coll.Watch(ctx, pipelineData, opts)

	} else if dbName != "" { // Watching a database

		fmt.Println("Watching", dbName)
		db := client.Database(dbName)
		cur, err = db.Watch(ctx, pipelineData, opts)

	} else { // Watching all

		fmt.Println("Watching all")
		cur, err = client.Watch(ctx, pipelineData, opts)
	}

	if err != nil {
		return
	}

	defer cur.Close(ctx)

	// loop forever look change data
	for cur.Next(ctx) {
		data := bson.M{}
		cur.Decode(&data)
		cb(data)
	}

	return
}

// Query Get many object in DB
// Maximum result is 1000
func (m *Instance) QueryWithOptions(query interface{}, opt *options.FindOptions) *common.APIResponse {
	if opt == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "Query error:  FindOptions missing required",
		}
	}

	if *opt.Limit <= 0 {
		*opt.Limit = 1000
	}

	if *opt.Limit > 1000 {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "Query error:  FindOptions Limit can't be greater than 1000",
		}
	}
	// check col
	if m.col == nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: Collection " + m.ColName + " is not init.",
		}
	}

	// transform to bson
	converted, err := m.convertToBson(query)
	if err != nil {
		return &common.APIResponse{
			Status:  common.APIStatus.Error,
			Message: "DB error: QueryOne - Cannot convert object - " + err.Error(),
		}
	}

	result, err := m.col.Find(context.TODO(), converted, opt)

	if err != nil || result.Err() != nil {
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ".",
			ErrorCode: "NOT_FOUND",
		}
	}

	list := m.newList(int(*opt.Limit))
	err = result.All(context.TODO(), &list)
	_ = result.Close(context.TODO())
	if err != nil || reflect.ValueOf(list).Len() == 0 {
		return &common.APIResponse{
			Status:    common.APIStatus.NotFound,
			Message:   "Not found any matched " + m.ColName + ".",
			ErrorCode: "NOT_FOUND",
		}
	}

	return &common.APIResponse{
		Status:  common.APIStatus.Ok,
		Message: "Query " + m.ColName + " successfully.",
		Data:    list,
	}
}

