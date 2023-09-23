package app

import (
	"context"

	v "github.com/core-go/core/v10"
	"github.com/core-go/health"
	"github.com/core-go/log"
	mgo "github.com/core-go/mongo"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "go-service/internal/handler"
	. "go-service/internal/repository"
	. "go-service/internal/service"
)

type ApplicationContext struct {
	Health *health.Handler
	User   UserTransport
}

func NewApp(ctx context.Context, conf Config) (*ApplicationContext, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.Mongo.Uri))
	db := client.Database(conf.Mongo.Database)
	if err != nil {
		return nil, err
	}
	logError := log.LogError
	validator, err := v.NewValidator()
	if err != nil {
		return nil, err
	}

	userRepository := NewUserAdapter(db, BuildQuery)
	userService := NewUserUseCase(userRepository)
	userHandler := NewUserHandler(userService, validator.Validate, logError)

	mongoChecker := mgo.NewHealthChecker(db.Client())
	healthHandler := health.NewHandler(mongoChecker)

	return &ApplicationContext{
		Health: healthHandler,
		User:   userHandler,
	}, nil
}
