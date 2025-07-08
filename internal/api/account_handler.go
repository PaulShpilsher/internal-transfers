package api

import (
	"internal-transfers/internal/model"
	"internal-transfers/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/shopspring/decimal"
)

type AccountHandler struct {
	Service *services.AccountService
}

func NewAccountHandler(service *services.AccountService) *AccountHandler {
	return &AccountHandler{Service: service}
}

func (h *AccountHandler) CreateAccount(ctx iris.Context) {
	var req CreateAccountRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StopWithStatus(iris.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}

	balance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		ctx.StopWithError(iris.StatusBadRequest, err)
		return
	}
	account := model.Account{
		AccountID: req.AccountID,
		Balance:   balance,
	}
	if err := h.Service.CreateAccount(account); err != nil {
		ctx.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	ctx.StatusCode(iris.StatusCreated)
}
