package main

import (
	"internal-transfers/internal/api"
	"internal-transfers/internal/config"
	"internal-transfers/internal/db"
	"internal-transfers/internal/services"

	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kataras/iris/v12"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		println("Config error:", err.Error())
		os.Exit(1)
	}

	// Initialize database connection
	dbConn, err := db.NewDBConnectionFromDSN(cfg.DBUrl)
	if err != nil {
		println("Database connection error:", err.Error())
		os.Exit(1)
	}
	defer dbConn.Close()

	// Initialize repositories and services
	repo := db.NewAccountRepository(dbConn)
	service := services.NewAccountService(repo)
	handler := api.NewAccountHandler(service)

	// Create and configure the Iris application
	app := iris.New()

	api.RegisterRoutes(app, handler)

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":"+cfg.ServerPort, iris.WithoutInterruptHandler); err != nil {
			app.Logger().Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	app.Logger().Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		app.Logger().Fatalf("Server forced to shutdown: %v", err)
	}
}
