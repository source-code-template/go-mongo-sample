package app

import (
	"github.com/core-go/core/server"
	"github.com/core-go/log/zap"
	mid "github.com/core-go/middleware"
)

type Config struct {
	Server     server.ServerConf `mapstructure:"server"`
	Mongo      MongoConfig       `mapstructure:"mongo"`
	Log        log.Config        `mapstructure:"log"`
	MiddleWare mid.LogConfig     `mapstructure:"middleware"`
}

type MongoConfig struct {
	Uri      string `yaml:"uri" mapstructure:"uri" json:"uri,omitempty" gorm:"column:uri" bson:"uri,omitempty" dynamodbav:"uri,omitempty" firestore:"uri,omitempty"`
	Database string `yaml:"database" mapstructure:"database" json:"database,omitempty" gorm:"column:database" bson:"database,omitempty" dynamodbav:"database,omitempty" firestore:"database,omitempty"`
}
