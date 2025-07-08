package main

import (
	"internal-transfers/internal/api"
	"internal-transfers/internal/db"
	"internal-transfers/internal/services"

	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/kataras/iris/v12"
)

func main() {
	_ = godotenv.Load()

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	repo, _ := db.NewAccountRepository(dsn)
	service := services.NewAccountService(repo)
	handler := api.NewAccountHandler(service)

	app := iris.New()
	app.Post("/accounts", handler.CreateAccount)

	// Register GET /accounts/{id} with id as int64
	app.Get("/accounts/{id:uint64}", handler.GetAccount)
	app.Listen(":8080")
}
