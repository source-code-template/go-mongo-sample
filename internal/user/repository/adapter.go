package repository

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mgo "github.com/core-go/mongo"
	"github.com/core-go/search/mongo/query"

	"go-service/internal/user/model"
)

func NewUserAdapter(db *mongo.Database, buildQuery func(*model.UserFilter) (bson.D, bson.M)) *UserAdapter {
	userType := reflect.TypeOf(model.User{})
	bsonMap := mgo.MakeBsonMap(userType)
	if buildQuery == nil {
		build := query.UseQuery(userType)
		buildQuery = func(filter *model.UserFilter) (d bson.D, m bson.M) {
			return build(filter)
		}
	}
	return &UserAdapter{Collection: db.Collection("users"), Map: bsonMap}
}

type UserAdapter struct {
	Collection *mongo.Collection
	Map        map[string]string
	BuildQuery func(*model.UserFilter) bson.D
}

func (r *UserAdapter) All(ctx context.Context) ([]model.User, error) {
	filter := bson.M{}
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var users []model.User
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserAdapter) Load(ctx context.Context, id string) (*model.User, error) {
	filter := bson.M{"_id": id}
	res := r.Collection.FindOne(ctx, filter)
	if res.Err() != nil {
		if strings.Compare(fmt.Sprint(res.Err()), "mongo: no documents in result") == 0 {
			return nil, nil
		} else {
			return nil, res.Err()
		}
	}
	var user model.User
	err := res.Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserAdapter) Create(ctx context.Context, user *model.User) (int64, error) {
	_, err := r.Collection.InsertOne(ctx, user)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key error collection:") {
			if strings.Contains(errMsg, "dup key: { _id: ") {
				return 0, err
			} else {
				return -1, err
			}
		}
		return 0, err
	}
	return 1, nil
}

func (r *UserAdapter) Update(ctx context.Context, user *model.User) (int64, error) {
	filter := bson.M{"_id": user.Id}
	update := bson.M{"$set": user}
	res, err := r.Collection.UpdateOne(ctx, filter, update)
	return res.ModifiedCount, err
}

func (r *UserAdapter) Patch(ctx context.Context, user map[string]interface{}) (int64, error) {
	filter := mgo.BuildQueryByIdFromMap(user, "id")
	bson := mgo.MapToBson(user, r.Map)
	return mgo.PatchOne(ctx, r.Collection, bson, filter)
}

func (r *UserAdapter) Delete(ctx context.Context, id string) (int64, error) {
	filter := bson.M{"_id": id}
	res, err := r.Collection.DeleteOne(ctx, filter)
	if res == nil || err != nil {
		return 0, err
	}
	return res.DeletedCount, err
}

func (r *UserAdapter) Search(ctx context.Context, filter *model.UserFilter, limit int64, offset int64) ([]model.User, int64, error) {
	query, fields := BuildQuery(filter)
	opts := options.Find()
	if len(filter.Sort) > 0 {
		opts.SetSort(mgo.BuildSort(filter.Sort, reflect.TypeOf(model.UserFilter{})))
	}
	opts.SetSkip(offset)
	if filter.Limit > 0 {
		opts.SetLimit(limit)
	}
	if fields != nil {
		opts.Projection = fields
	}

	var users []model.User
	cursor, err := r.Collection.Find(ctx, query, opts)
	if err != nil {
		if strings.Contains(err.Error(), "mongo: no documents in result") {
			return users, 0, nil
		}
		return users, 0, err
	}

	err = cursor.All(ctx, &users)
	if err != nil {
		return users, 0, err
	}
	total, err := r.Collection.CountDocuments(ctx, query)
	return users, total, err
}

func BuildQuery(filter *model.UserFilter) (bson.D, bson.M) {
	query := bson.D{}
	if len(filter.Id) > 0 {
		query = append(query, bson.E{Key: "_id", Value: filter.Id})
	}
	if len(filter.Username) > 0 {
		query = append(query, bson.E{Key: "username", Value: primitive.Regex{Pattern: fmt.Sprintf("^%v", filter.Username), Options: "i"}})
	}
	if len(filter.Email) > 0 {
		query = append(query, bson.E{Key: "email", Value: primitive.Regex{Pattern: fmt.Sprintf("^%v", filter.Email), Options: "i"}})
	}
	if len(filter.Phone) > 0 {
		query = append(query, bson.E{Key: "phone", Value: primitive.Regex{Pattern: fmt.Sprintf("\\w*%v\\w*", filter.Phone), Options: "i"}})
	}

	fields := mgo.GetFields(filter.Fields, reflect.TypeOf(model.User{}))
	return query, fields
}
