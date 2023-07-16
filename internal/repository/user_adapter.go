package repository

import (
	"context"
	"fmt"
	mgo "github.com/core-go/mongo"
	s "github.com/core-go/search"
	"github.com/core-go/search/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"strings"

	. "go-service/internal/filter"
	. "go-service/internal/model"
)

func NewUserAdapter(db *mongo.Database, buildQuery func(*UserFilter) (bson.D, bson.M)) *UserAdapter {
	userType := reflect.TypeOf(User{})
	bsonMap := mgo.MakeBsonMap(userType)
	if buildQuery == nil {
		build := query.UseQuery(userType)
		buildQuery = func(filter *UserFilter) (d bson.D, m bson.M) {
			return build(filter)
		}
	}
	return &UserAdapter{Collection: db.Collection("users"), Map: bsonMap}
}

type UserAdapter struct {
	Collection *mongo.Collection
	Map        map[string]string
	BuildQuery func(*UserFilter) bson.D
}

func (r *UserAdapter) All(ctx context.Context) ([]User, error) {
	filter := bson.M{}
	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var users []User
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserAdapter) Load(ctx context.Context, id string) (*User, error) {
	filter := bson.M{"_id": id}
	res := r.Collection.FindOne(ctx, filter)
	if res.Err() != nil {
		if strings.Compare(fmt.Sprint(res.Err()), "mongo: no documents in result") == 0 {
			return nil, nil
		} else {
			return nil, res.Err()
		}
	}
	var user User
	err := res.Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserAdapter) Create(ctx context.Context, user *User) (int64, error) {
	_, err := r.Collection.InsertOne(ctx, user)
	if err != nil {
		errMsg := err.Error()
		if strings.Index(errMsg, "duplicate key error collection:") >= 0 {
			if strings.Index(errMsg, "dup key: { _id: ") >= 0 {
				return 0, nil
			} else {
				return -1, nil
			}
		}
		return 0, err
	}
	return 1, nil
}

func (r *UserAdapter) Update(ctx context.Context, user *User) (int64, error) {
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
func (r *UserAdapter) Search(ctx context.Context, filter *UserFilter) ([]User, int64, error) {
	query, fields := BuildQuery(filter)
	opts := options.Find()
	if len(filter.Sort) > 0 {
		opts.SetSort(mgo.BuildSort(filter.Sort, reflect.TypeOf(UserFilter{})))
	}
	offset := s.GetOffset(filter.Limit, filter.Page)
	opts.SetSkip(offset)
	if filter.Limit > 0 {
		opts.SetLimit(filter.Limit)
	}
	if fields != nil {
		opts.Projection = fields
	}

	var users []User
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

func BuildQuery(filter *UserFilter) (bson.D, bson.M) {
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

	userType := reflect.TypeOf(User{})
	if len(filter.Fields) > 0 {
		var fields = bson.M{}
		for _, key := range filter.Fields {
			_, _, columnName := mgo.GetFieldByJson(userType, key)
			if len(columnName) < 0 {
				return query, nil
			}
			fields[columnName] = 1
		}
		return query, fields
	}
	return query, nil
}
