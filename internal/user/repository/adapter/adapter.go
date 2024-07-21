package adapter

import (
	"context"
	"errors"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mgo "github.com/core-go/mongo"
	"go-service/internal/user/model"
)

func NewUserAdapter(db *mongo.Database, buildQuery func(*model.UserFilter) (bson.D, bson.M)) *UserAdapter {
	userType := reflect.TypeOf(model.User{})
	bsonMap := mgo.MakeBsonMap(userType)
	return &UserAdapter{Collection: db.Collection("users"), Map: bsonMap, BuildQuery: buildQuery}
}

type UserAdapter struct {
	Collection *mongo.Collection
	Map        map[string]string
	BuildQuery func(*model.UserFilter) (bson.D, bson.M)
}

func (r *UserAdapter) All(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := mgo.Find(ctx, r.Collection, bson.D{}, &users)
	return users, err
}

func (r *UserAdapter) Load(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	filter := bson.M{"_id": id}
	ok, err := mgo.FindOne(ctx, r.Collection, filter, &user)
	if !ok || err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserAdapter) Create(ctx context.Context, user *model.User) (int64, error) {
	_, res, err := mgo.InsertOne(ctx, r.Collection, user)
	return res, err
}

func (r *UserAdapter) Update(ctx context.Context, user *model.User) (int64, error) {
	return mgo.UpdateOne(ctx, r.Collection, user.Id, user)
}

func (r *UserAdapter) Patch(ctx context.Context, user map[string]interface{}) (int64, error) {
	id, ok := user["id"]
	if !ok {
		return -1, errors.New("id must be in map[string]interface{} for patch")
	}
	bsonObj := mgo.MapToBson(user, r.Map)
	return mgo.PatchOne(ctx, r.Collection, id, bsonObj)
}

func (r *UserAdapter) Delete(ctx context.Context, id string) (int64, error) {
	return mgo.DeleteOne(ctx, r.Collection, id)
}

func (r *UserAdapter) Search(ctx context.Context, filter *model.UserFilter, limit int64, offset int64) ([]model.User, int64, error) {
	query, fields := r.BuildQuery(filter)
	var users []model.User
	total, err := r.Collection.CountDocuments(ctx, query)
	if err != nil || total == 0 {
		return users, total, err
	}
	opts := options.Find()
	if len(filter.Sort) > 0 {
		opts.SetSort(mgo.BuildSort(filter.Sort, reflect.TypeOf(model.UserFilter{})))
	}
	opts.SetSkip(offset)
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if fields != nil {
		opts.Projection = fields
	}
	cursor, err := r.Collection.Find(ctx, query, opts)
	if err != nil {
		return users, total, err
	}
	err = cursor.All(ctx, &users)
	return users, total, err
}
