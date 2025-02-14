package main

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"testAlvtoShp/internal/config"
	"testAlvtoShp/internal/handler"
	"testAlvtoShp/internal/logger"
	"testAlvtoShp/internal/server"
	"testAlvtoShp/internal/service"
	"testAlvtoShp/internal/store"
)

// @title Avito SHop API
// @version 1.0
// @description API Server 4 Test Avito

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	ctx := context.Background()
	log := logger.InitLogger()
	ctx = logger.ContextWithLogger(ctx, log)

	cfg := config.InitConfig(ctx)
	dbConn, err := store.NewDbConn(ctx, cfg)
	if err != nil {
		log.Fatalw("error with connecting to db", zap.Error(err))
		return
	}

	storeLevel := store.NewStore(dbConn)
	serviceLevel := service.NewService(storeLevel, cfg)
	handlerLevel := handler.NewHandler(serviceLevel)

	httpServer := new(server.Server)

	go func() {
		if err = httpServer.InitServer(cfg.Port, handlerLevel.InitRoutes(ctx)); err != nil {
			log.Fatalw("error with initializing server", zap.Error(err))
			return
		}
	}()

	log.Infow("starting server in port", "port", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Infow("shutting down server in port", "port", cfg.Port)

	if err = httpServer.Shutdown(ctx); err != nil {
		log.Fatalw("error with shutting down server", zap.Error(err))
		os.Exit(1)
	}

	if err = store.ShutDown(ctx, dbConn); err != nil {
		log.Fatalw("error with shutting down store", zap.Error(err))
		os.Exit(1)
	}
}
