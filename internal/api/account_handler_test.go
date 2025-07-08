package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"internal-transfers/internal/mocks"
	"internal-transfers/internal/model"

	"github.com/golang/mock/gomock"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/httptest"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func setupTestApp(_ *testing.T, mockSvc *mocks.MockAccountServicePort) *iris.Application {
	handler := NewAccountHandler(mockSvc)
	app := iris.New()
	RegisterRoutes(app, handler)
	return app
}

func TestCreateAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	req := CreateAccountRequest{
		AccountID:      123,
		InitialBalance: "100.50",
	}
	acc := model.Account{
		AccountID: req.AccountID,
		Balance:   decimal.RequireFromString(req.InitialBalance),
	}
	mockSvc.EXPECT().CreateAccount(acc).Return(nil)
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/accounts").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusCreated)
}

func TestCreateAccount_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	resp := httptest.New(t, app).POST("/accounts").WithHeader("Content-Type", "application/json").WithText("not-json").Expect()
	resp.Status(http.StatusBadRequest)
	resp.JSON().Object().ValueEqual("error", "invalid request body: invalid character 'o' in literal null (expecting 'u')")
}

func TestCreateAccount_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	// AccountID is 0 (invalid)
	req := CreateAccountRequest{AccountID: 0, InitialBalance: "100.00"}
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/accounts").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusBadRequest)
	resp.JSON().Object().Value("error").String().Contains("validation error")
}

func TestCreateAccount_InvalidInitialBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	req := CreateAccountRequest{AccountID: 1, InitialBalance: "not-a-number"}
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/accounts").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusBadRequest)
	resp.JSON().Object().Value("error").String().Contains("invalid initial balance")
}

func TestCreateAccount_ServiceValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)

	testCases := []struct {
		name string
		err  error
	}{
		{"AccountIDMustBePositive", model.ErrAccountIDMustBePositive},
		{"BalanceMustBeNonNegative", model.ErrBalanceMustBeNonNegative},
		{"PrecisionTooHigh", model.ErrPrecisionTooHigh},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateAccountRequest{AccountID: 1, InitialBalance: "100.00"}
			acc := model.Account{AccountID: req.AccountID, Balance: decimal.RequireFromString(req.InitialBalance)}
			mockSvc.EXPECT().CreateAccount(acc).Return(tc.err)
			body, _ := json.Marshal(req)
			resp := httptest.New(t, app).POST("/accounts").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
			resp.Status(http.StatusBadRequest)
			resp.JSON().Object().Value("error").String().Contains(tc.err.Error())
		})
	}
}

func TestCreateAccount_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	req := CreateAccountRequest{AccountID: 1, InitialBalance: "100.00"}
	acc := model.Account{AccountID: req.AccountID, Balance: decimal.RequireFromString(req.InitialBalance)}
	mockSvc.EXPECT().CreateAccount(acc).Return(model.ErrAccountIDAlreadyExists)
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/accounts").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusConflict)
	resp.JSON().Object().Value("error").String().Contains(model.ErrAccountIDAlreadyExists.Error())
}

func TestCreateAccount_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	req := CreateAccountRequest{AccountID: 1, InitialBalance: "100.00"}
	acc := model.Account{AccountID: req.AccountID, Balance: decimal.RequireFromString(req.InitialBalance)}
	mockSvc.EXPECT().CreateAccount(acc).Return(assert.AnError)
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/accounts").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusInternalServerError)
	resp.JSON().Object().Value("error").String().Contains("failed to create account")
}

func TestGetAccount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	acc := model.Account{AccountID: 42, Balance: decimal.NewFromFloat(123.45)}
	mockSvc.EXPECT().GetAccount(acc.AccountID).Return(acc, nil)
	resp := httptest.New(t, app).GET("/accounts/42").Expect()
	resp.Status(http.StatusOK)
	resp.JSON().Object().ValueEqual("account_id", float64(acc.AccountID))
	resp.JSON().Object().ValueEqual("balance", acc.Balance.String())
}

func TestGetAccount_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	resp := httptest.New(t, app).GET("/accounts/notanumber").Expect()
	resp.Status(http.StatusNotFound)
	// No JSON body is expected, as Iris returns 404 with empty body for param type mismatch
}

func TestGetAccount_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	mockSvc.EXPECT().GetAccount(int64(404)).Return(model.Account{}, model.ErrAccountNotFound)
	resp := httptest.New(t, app).GET("/accounts/404").Expect()
	resp.Status(http.StatusNotFound)
	resp.JSON().Object().Value("error").String().Contains("account not found")
}

func TestGetAccount_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	mockSvc.EXPECT().GetAccount(int64(500)).Return(model.Account{}, assert.AnError)
	resp := httptest.New(t, app).GET("/accounts/500").Expect()
	resp.Status(http.StatusInternalServerError)
	resp.JSON().Object().Value("error").String().Contains("internal server error")
}

func TestGetAccount_ResponseWriteError(t *testing.T) {
	// Iris httptest does not simulate JSON write errors, so this is not testable in this context.
}

func TestSubmitTransaction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	req := CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "10.00"}
	amount := decimal.RequireFromString(req.Amount)
	mockSvc.EXPECT().Transfer(req.SourceAccountID, req.DestinationAccountID, amount).Return(nil)
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusOK)
}

func TestSubmitTransaction_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithText("not-json").Expect()
	resp.Status(http.StatusBadRequest)
	resp.JSON().Object().Value("error").String().Contains("invalid request body")
}

func TestSubmitTransaction_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	req := CreateTransactionRequest{SourceAccountID: 0, DestinationAccountID: 2, Amount: "10.00"}
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusBadRequest)
	resp.JSON().Object().Value("error").String().Contains("validation error")
}

func TestSubmitTransaction_InvalidAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	req := CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "not-a-number"}
	body, _ := json.Marshal(req)
	resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusBadRequest)
	resp.JSON().Object().Value("error").String().Contains("invalid amount")
}

func TestSubmitTransaction_ServiceValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)

	testCases := []struct {
		name string
		err  error
	}{
		{"AccountIDMustBePositive", model.ErrAccountIDMustBePositive},
		{"SourceAndDestinationMustDiffer", model.ErrSourceAndDestinationMustDiffer},
		{"AmountMustBePositive", model.ErrAmountMustBePositive},
		{"PrecisionTooHigh", model.ErrPrecisionTooHigh},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "10.00"}
			amount := decimal.RequireFromString(req.Amount)
			mockSvc.EXPECT().Transfer(req.SourceAccountID, req.DestinationAccountID, amount).Return(tc.err)
			body, _ := json.Marshal(req)
			resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
			resp.Status(http.StatusBadRequest)
			resp.JSON().Object().Value("error").String().Contains(tc.err.Error())
		})
	}
}

func TestSubmitTransaction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)

	testCases := []error{model.ErrSourceAccountNotFound, model.ErrDestinationAccountNotFound}
	for _, errVal := range testCases {
		mockSvc.EXPECT().Transfer(int64(1), int64(2), decimal.RequireFromString("10.00")).Return(errVal)
		body, _ := json.Marshal(CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "10.00"})
		resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
		resp.Status(http.StatusNotFound)
		resp.JSON().Object().Value("error").String().Contains(errVal.Error())
	}
}

func TestSubmitTransaction_InsufficientFunds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	mockSvc.EXPECT().Transfer(int64(1), int64(2), decimal.RequireFromString("10.00")).Return(model.ErrInsufficientFunds)
	body, _ := json.Marshal(CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "10.00"})
	resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusBadRequest)
	resp.JSON().Object().Value("error").String().Contains(model.ErrInsufficientFunds.Error())
}

func TestSubmitTransaction_InternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := mocks.NewMockAccountServicePort(ctrl)
	app := setupTestApp(t, mockSvc)
	mockSvc.EXPECT().Transfer(int64(1), int64(2), decimal.RequireFromString("10.00")).Return(assert.AnError)
	body, _ := json.Marshal(CreateTransactionRequest{SourceAccountID: 1, DestinationAccountID: 2, Amount: "10.00"})
	resp := httptest.New(t, app).POST("/transactions").WithHeader("Content-Type", "application/json").WithBytes(body).Expect()
	resp.Status(http.StatusInternalServerError)
	resp.JSON().Object().Value("error").String().Contains("failed to submit transaction")
}
