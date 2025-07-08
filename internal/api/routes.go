package api

import (
	"log"

	"github.com/kataras/iris/v12"
)

func RegisterRoutes(app *iris.Application, handler *AccountHandler) {
	// Global error handler middleware
	app.Use(func(ctx iris.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC: %v", r)
				ctx.StatusCode(iris.StatusInternalServerError)
				ctx.JSON(ErrorResponse{Error: "internal server error"})
			}
		}()
		ctx.Next()
	})

	app.Post("/accounts", handler.CreateAccount)
	app.Get("/accounts/{id:uint64}", handler.GetAccount)
	app.Post("/transactions", handler.SubmitTransaction)
}
