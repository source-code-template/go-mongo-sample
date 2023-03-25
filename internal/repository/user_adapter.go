package repository

import (
	"context"
	"fmt"
	mgo "github.com/core-go/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"strings"

	. "go-service/internal/model"
)

func NewUserAdapter(db *mongo.Database) *UserAdapter {
	userType := reflect.TypeOf(User{})
	bsonMap := mgo.MakeBsonMap(userType)
	return &UserAdapter{Collection: db.Collection("users"), Map: bsonMap}
}

type UserAdapter struct {
	Collection *mongo.Collection
	Map        map[string]string
}

func (r *UserAdapter) All(ctx context.Context) (*[]User, error) {
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
	return &users, nil
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
