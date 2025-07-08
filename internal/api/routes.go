package api

import (
	"github.com/kataras/iris/v12"
)

func RegisterRoutes(app *iris.Application, handler *AccountHandler) {
	app.Post("/accounts", handler.CreateAccount)
	app.Get("/accounts/{id:uint64}", handler.GetAccount)
	app.Post("/transactions", handler.SubmitTransaction)
}
