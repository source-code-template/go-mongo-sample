package main

import (
	"context"

	"github.com/core-go/config"
	srv "github.com/core-go/core/server"
	"github.com/core-go/log/zap"
	mid "github.com/core-go/middleware"
	"github.com/gorilla/mux"

	"go-service/internal/app"
)

func main() {
	var cfg app.Config
	err := config.Load(&cfg, "configs/config")
	if err != nil {
		panic(err)
	}
	r := mux.NewRouter()

	log.Initialize(cfg.Log)
	r.Use(mid.BuildContext)
	logger := mid.NewLogger()
	if log.IsInfoEnable() {
		r.Use(mid.Logger(cfg.MiddleWare, log.InfoFields, logger))
	}
	r.Use(mid.Recover(log.PanicMsg))

	ctx := context.Background()
	err = app.Route(ctx, r, cfg)
	if err != nil {
		panic(err)
	}
	log.Info(ctx, srv.ServerInfo(cfg.Server))
	server := srv.CreateServer(cfg.Server, r)
	if err = server.ListenAndServe(); err != nil {
		log.Error(ctx, err.Error())
	}
}
