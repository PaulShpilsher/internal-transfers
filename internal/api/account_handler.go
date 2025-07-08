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

type ErrorResponse struct {
	Error string `json:"error"`
}

type AccountHandler struct {
	Service *services.AccountService
}

func NewAccountHandler(service *services.AccountService) *AccountHandler {
	return &AccountHandler{Service: service}
}

func (h *AccountHandler) CreateAccount(ctx iris.Context) {
	var req CreateAccountRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(ErrorResponse{Error: "validation error: " + err.Error()})
		return
	}

	balance, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(ErrorResponse{Error: "invalid initial balance: " + err.Error()})
		return
	}
	account := model.Account{
		AccountID: req.AccountID,
		Balance:   balance,
	}
	if err := h.Service.CreateAccount(account); err != nil {
		switch {
		case errors.Is(err, model.ErrAccountIDMustBePositive),
			errors.Is(err, model.ErrBalanceMustBeNonNegative),
			errors.Is(err, model.ErrPrecisionTooHigh):
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(ErrorResponse{Error: err.Error()})
			return
		case errors.Is(err, model.ErrAccountIDAlreadyExists):
			ctx.StatusCode(iris.StatusConflict)
			ctx.JSON(ErrorResponse{Error: err.Error()})
			return
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(ErrorResponse{Error: "failed to create account: " + err.Error()})
			return
		}
	}
	ctx.StatusCode(iris.StatusCreated)
}

func (h *AccountHandler) GetAccount(ctx iris.Context) {
	// Example: get ID from URL
	idStr := ctx.URLParam("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(ErrorResponse{Error: "invalid account id: " + err.Error()})
		return
	}

	account, err := h.Service.GetAccount(id)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			ctx.StatusCode(iris.StatusNotFound)
			ctx.JSON(ErrorResponse{Error: "account not found"})
			return
		}
		log.Printf("get account error: %v", err)
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(ErrorResponse{Error: "internal server error"})
		return
	}

	resp := GetAccountResponse{
		AccountID: account.AccountID,
		Balance:   account.Balance.String(),
	}
	if err := ctx.JSON(resp); err != nil {
		log.Printf("failed to write response: %v", err)
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.JSON(ErrorResponse{Error: "internal server error"})
		return
	}
}

func (h *AccountHandler) SubmitTransaction(ctx iris.Context) {
	var req CreateTransactionRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(ErrorResponse{Error: "invalid request body: " + err.Error()})
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(ErrorResponse{Error: "validation error: " + err.Error()})
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(ErrorResponse{Error: "invalid amount: " + err.Error()})
		return
	}

	err = h.Service.Transfer(req.SourceAccountID, req.DestinationAccountID, amount)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrAccountIDMustBePositive),
			errors.Is(err, model.ErrSourceAndDestinationMustDiffer),
			errors.Is(err, model.ErrAmountMustBePositive),
			errors.Is(err, model.ErrPrecisionTooHigh):
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(ErrorResponse{Error: err.Error()})
			return
		case errors.Is(err, model.ErrSourceAccountNotFound), errors.Is(err, model.ErrDestinationAccountNotFound):
			ctx.StatusCode(iris.StatusNotFound)
			ctx.JSON(ErrorResponse{Error: err.Error()})
			return
		case errors.Is(err, model.ErrInsufficientFunds):
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(ErrorResponse{Error: err.Error()})
			return
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			ctx.JSON(ErrorResponse{Error: "failed to submit transaction: " + err.Error()})
			return
		}
	}
	ctx.StatusCode(iris.StatusOK)
}
