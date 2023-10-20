package app

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taraslis453/shopify-customer-auth/config"
	"github.com/taraslis453/shopify-customer-auth/pkg/httpserver"
	"github.com/taraslis453/shopify-customer-auth/pkg/logging"
	"github.com/taraslis453/shopify-customer-auth/pkg/postgresql"

	httpController "github.com/taraslis453/shopify-customer-auth/internal/controller/http"
	"github.com/taraslis453/shopify-customer-auth/internal/entity"
	"github.com/taraslis453/shopify-customer-auth/internal/service"
)

// Run - initializes and runs application.
func Run(cfg *config.Config) {
	logger := logging.NewZapLogger(cfg.Log.Level)

	postgresql, err := postgresql.NewPostgreSQLGorm(postgresql.Config{
		User:     cfg.PostgreSQL.User,
		Password: cfg.PostgreSQL.Password,
		Host:     cfg.PostgreSQL.Host,
		Database: cfg.PostgreSQL.Database,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to init repository: %w", err))
	}

	err = postgresql.DB.AutoMigrate(
		&entity.User{},
	)
	if err != nil {
		log.Fatal(fmt.Errorf("automigration failed: %w", err))
	}

	storages := service.Storages{}

	serviceOptions := service.Options{
		Storages: storages,
		Config:   cfg,
		Logger:   logger,
	}

	services := service.Services{
		Customer: service.NewCustomerService(serviceOptions),
	}

	httpHandler := gin.New()

	httpController.New(httpController.Options{
		Handler:  httpHandler,
		Services: services,
		Logger:   logger,
		Config:   cfg,
	})

	httpServer := httpserver.New(
		httpHandler,
		httpserver.Port(cfg.HTTP.Port),
		httpserver.ReadTimeout(time.Second*60),
		httpserver.WriteTimeout(time.Second*60),
		httpserver.ShutdownTimeout(time.Second*30),
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logger.Info("app - Run - signal: " + s.String())

	case err = <-httpServer.Notify():
		logger.Error("app - Run - httpServer.Notify", "err", err)
	}

	err = httpServer.Shutdown()
	if err != nil {
		logger.Error("app - Run - httpServer.Shutdown", "err", err)
	}
}
