package api

import (
	"internal-transfers/internal/model"
	"internal-transfers/internal/services"

	"errors"
	"log"
	"strconv"

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

func (h *AccountHandler) GetAccount(ctx iris.Context) {
	// Example: get ID from URL
	idStr := ctx.URLParam("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.StopWithError(iris.StatusBadRequest, errors.New("invalid account id"))
		return
	}

	account, err := h.Service.GetAccount(id)
	if errors.Is(err, model.ErrAccountNotFound) {
		ctx.StopWithError(iris.StatusNotFound, errors.New("account not found"))
		return
	}
	if err != nil {
		log.Printf("get account error: %v", err)
		ctx.StopWithError(iris.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	resp := GetAccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}
	if err := ctx.JSON(resp); err != nil {
		log.Printf("failed to write response: %v", err)
		ctx.StopWithError(iris.StatusInternalServerError, errors.New("internal server error"))
		return
	}
}
