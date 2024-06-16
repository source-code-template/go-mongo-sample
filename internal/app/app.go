package app

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/core-go/health"
	"github.com/core-go/log/zap"
	mgo "github.com/core-go/mongo/health"

	"go-service/internal/user"
)

type ApplicationContext struct {
	Health *health.Handler
	User   user.UserTransport
}

func NewApp(ctx context.Context, cfg Config) (*ApplicationContext, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.Uri))
	db := client.Database(cfg.Mongo.Database)
	if err != nil {
		return nil, err
	}
	logError := log.LogError

	userHandler, err := user.NewUserHandler(db, logError)
	if err != nil {
		return nil, err
	}

	mongoChecker := mgo.NewHealthChecker(db.Client())
	healthHandler := health.NewHandler(mongoChecker)

	return &ApplicationContext{
		Health: healthHandler,
		User:   userHandler,
	}, nil
}
