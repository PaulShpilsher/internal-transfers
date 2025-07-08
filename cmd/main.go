package main

import (
	"internal-transfers/internal/api"
	"internal-transfers/internal/db"
	"internal-transfers/internal/services"

	"github.com/kataras/iris/v12"
)

func main() {
	app := iris.New()
	repo, _ := db.NewAccountRepository("your-dsn")
	service := services.NewAccountService(repo)
	handler := api.NewAccountHandler(service)

	app.Post("/accounts", handler.CreateAccount)
	app.Listen(":8080")
}
