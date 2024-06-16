package user

import (
	"context"
	"net/http"

	v "github.com/core-go/core/v10"
	"go.mongodb.org/mongo-driver/mongo"

	"go-service/internal/user/handler"
	"go-service/internal/user/repository"
	"go-service/internal/user/service"
)

type UserTransport interface {
	All(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)
	Load(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Patch(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

func NewUserHandler(db *mongo.Database, logError func(context.Context, string, ...map[string]interface{})) (UserTransport, error) {
	validator, err := v.NewValidator()
	if err != nil {
		return nil, err
	}

	userRepository := repository.NewUserAdapter(db, repository.BuildQuery)
	// userRepository := adapter.NewSearchAdapterWithVersion[model.User, string, *model.UserFilter]()
	userService := service.NewUserUseCase(userRepository)
	userHandler := handler.NewUserHandler(userService, validator.Validate, logError)
	return userHandler, nil
}
