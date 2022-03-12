# go-mongo-layer-architecture-sample

#### To run the application
```shell
go run main.go
```

## API Design
### Common HTTP methods
- GET: retrieve a representation of the resource
- POST: create a new resource
- PUT: update the resource
- PATCH: perform a partial update of a resource, refer to [service](https://github.com/core-go/service) and [mongo](https://github.com/core-go/mongo)  
- DELETE: delete a resource

## API design for health check
To check if the service is available.
#### *Request:* GET /health
#### *Response:*
```json
{
    "status": "UP",
    "details": {
        "mongo": {
            "status": "UP"
        }
    }
}
```

## API design for users
#### *Resource:* users

### Get all users
#### *Request:* GET /users
#### *Response:*
```json
[
    {
        "id": "spiderman",
        "username": "peter.parker",
        "email": "peter.parker@gmail.com",
        "phone": "0987654321",
        "dateOfBirth": "1962-08-25T16:59:59.999Z"
    },
    {
        "id": "wolverine",
        "username": "james.howlett",
        "email": "james.howlett@gmail.com",
        "phone": "0987654321",
        "dateOfBirth": "1974-11-16T16:59:59.999Z"
    }
]
```

### Get one user by id
#### *Request:* GET /users/:id
```shell
GET /users/wolverine
```
#### *Response:*
```json
{
    "id": "wolverine",
    "username": "james.howlett",
    "email": "james.howlett@gmail.com",
    "phone": "0987654321",
    "dateOfBirth": "1974-11-16T16:59:59.999Z"
}
```

### Create a new user
#### *Request:* POST /users 
```json
{
    "id": "wolverine",
    "username": "james.howlett",
    "email": "james.howlett@gmail.com",
    "phone": "0987654321",
    "dateOfBirth": "1974-11-16T16:59:59.999Z"
}
```
#### *Response:* 1: success, 0: duplicate key, -1: error
```json
1
```

### Update one user by id
#### *Request:* PUT /users/:id
```shell
PUT /users/wolverine
```
```json
{
    "username": "james.howlett",
    "email": "james.howlett@gmail.com",
    "phone": "0987654321",
    "dateOfBirth": "1974-11-16T16:59:59.999Z"
}
```
#### *Response:* 1: success, 0: not found, -1: error
```json
1
```

### Patch one user by id
Perform a partial update of user. For example, if you want to update 2 fields: email and phone, you can send the request body of below.
#### *Request:* PATCH /users/:id
```shell
PATCH /users/wolverine
```
```json
{
    "email": "james.howlett@gmail.com",
    "phone": "0987654321"
}
```
#### *Response:* 1: success, 0: not found, -1: error
```json
1
```

#### Problems for patch
If we pass a struct as a parameter, we cannot control what fields we need to update. So, we must pass a map as a parameter.
```go
type UserService interface {
    Update(ctx context.Context, user *User) (int64, error)
    Patch(ctx context.Context, user map[string]interface{}) (int64, error)
}
```
We must solve 2 problems:
1. At http handler layer, we must convert the user struct to map, with json format, and make sure the nested data types are passed correctly.
2. At service layer or repository layer, from json format, we must convert the json format to database format (in this case, we must convert to bson of Mongo)

#### Solutions for patch  
1. At http handler layer, we use [core-go/service](https://github.com/core-go/service), to convert the user struct to map, to make sure we just update the fields we need to update
```go
import server "github.com/core-go/service"

func (h *UserHandler) Patch(w http.ResponseWriter, r *http.Request) {
    var user User
    userType := reflect.TypeOf(user)
    _, jsonMap := sv.BuildMapField(userType)
    body, _ := sv.BuildMapAndStruct(r, &user)
    json, er1 := sv.BodyToJson(r, user, body, ids, jsonMap, nil)

    result, er2 := h.service.Patch(r.Context(), json)
    if er2 != nil {
        http.Error(w, er2.Error(), http.StatusInternalServerError)
        return
    }
    respond(w, result)
}
```

2. At service layer or repository layer, we use [core-go/mongo](https://github.com/core-go/mongo), to convert from json to bson 
```go
import mgo "github.com/core-go/mongo"

func (p *MongoUserService) Patch(ctx context.Context, user map[string]interface{}) (int64, error) {
    userType := reflect.TypeOf(User{})
    maps := mgo.MakeBsonMap(userType)
    filter := mgo.BuildQueryByIdFromMap(user, "id")
    bson := mgo.MapToBson(user, maps)
    return mgo.PatchOne(ctx, p.Collection, bson, filter)
}
```

### Delete a new user by id
#### *Request:* DELETE /users/:id
```shell
DELETE /users/wolverine
```
#### *Response:* 1: success, 0: not found, -1: error
```json
1
```

## Common libraries
- [core-go/health](https://github.com/core-go/health): include HealthHandler, HealthChecker, MongoHealthChecker
- [core-go/config](https://github.com/core-go/config): to load the config file, and merge with other environments (SIT, UAT, ENV)
- [core-go/log](https://github.com/core-go/log): log and log middleware

### core-go/health
To check if the service is available, refer to [core-go/health](https://github.com/core-go/health)
#### *Request:* GET /health
#### *Response:*
```json
{
    "status": "UP",
    "details": {
        "mongo": {
            "status": "UP"
        }
    }
}
```
To create health checker, and health handler
```go
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(root.Mongo.Uri))
    if err != nil {
        return nil, err
    }
    db := client.Database(root.Mongo.Database)

    mongoChecker := mongo.NewHealthChecker(db)
    healthHandler := health.NewHealthHandler(mongoChecker)
```

To handler routing
```go
    r := mux.NewRouter()
    r.HandleFunc("/health", healthHandler.Check).Methods("GET")
```

### core-go/config
To load the config from "config.yml", in "configs" folder
```go
package main

import "github.com/core-go/config"

type Root struct {
    DB DatabaseConfig `mapstructure:"db"`
}

type DatabaseConfig struct {
    Driver                 string `mapstructure:"driver"`
    DataSourceName string `mapstructure:"data_source_name"`
}

func main() {
    var conf Root
    err := config.Load(&conf, "configs/config")
    if err != nil {
        panic(err)
    }
}
```

### core-go/log *&* core-go/log/middleware
```go
import (
	"github.com/core-go/config"
	"github.com/core-go/log"
	mid "github.com/core-go/log/middleware"
	"github.com/gorilla/mux"
)

func main() {
	var conf app.Root
	config.Load(&conf, "configs/config")

	r := mux.NewRouter()

	log.Initialize(conf.Log)
	r.Use(mid.BuildContext)
	logger := mid.NewLogger()
	r.Use(mid.Logger(conf.MiddleWare, log.InfoFields, logger))
	r.Use(mid.Recover(log.ErrorMsg))
}
```
To configure to ignore the health check, use "skips":
```yaml
middleware:
  skips: /health
```