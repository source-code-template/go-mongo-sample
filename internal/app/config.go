package app

import (
	"github.com/core-go/core/server"
	"github.com/core-go/log/zap"
	mid "github.com/core-go/middleware"
	"github.com/core-go/mongo/client"
)

type Config struct {
	Server     server.ServerConf  `mapstructure:"server"`
	Mongo      client.MongoConfig `mapstructure:"mongo"`
	Log        log.Config         `mapstructure:"log"`
	MiddleWare mid.LogConfig      `mapstructure:"middleware"`
}
