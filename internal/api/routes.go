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

	// Middleware to restrict POST to application/json and limit body size to 4KB
	jsonAndSizeLimit := func(ctx iris.Context) {
		if ctx.Method() == iris.MethodPost {
			if ctx.GetHeader("Content-Type") != "application/json" {
				ctx.StatusCode(iris.StatusUnsupportedMediaType)
				ctx.JSON(ErrorResponse{Error: "Content-Type must be application/json"})
				return
			}
			ctx.SetMaxRequestBodySize(4096) // 4KB
		}
		ctx.Next()
	}

	app.Post("/accounts", jsonAndSizeLimit, handler.CreateAccount)
	app.Get("/accounts/{id:uint64}", handler.GetAccount)
	app.Post("/transactions", jsonAndSizeLimit, handler.SubmitTransaction)
}
