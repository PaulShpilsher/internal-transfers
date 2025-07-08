package services

import (
	"errors"
	"testing"

	"internal-transfers/internal/mocks"
	"internal-transfers/internal/model"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func validAccount() model.Account {
	return model.Account{
		AccountID: 1,
		Balance:   decimal.NewFromFloat(100.0),
	}
}

func TestCreateAccount_ValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAccountRepositoryPort(ctrl)
	svc := NewAccountService(repo)

	acc := validAccount()
	acc.AccountID = 0
	err := svc.CreateAccount(acc)
	assert.ErrorIs(t, err, model.ErrAccountIDMustBePositive)

	acc = validAccount()
	acc.Balance = decimal.NewFromFloat(-1)
	err = svc.CreateAccount(acc)
	assert.ErrorIs(t, err, model.ErrBalanceMustBeNonNegative)
}

func TestCreateAccount_RepositoryErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAccountRepositoryPort(ctrl)
	svc := NewAccountService(repo)

	acc := validAccount()
	repo.EXPECT().CreateAccount(acc.AccountID, acc.Balance).Return(model.ErrAccountIDAlreadyExists)
	err := svc.CreateAccount(acc)
	assert.ErrorIs(t, err, model.ErrAccountIDAlreadyExists)
}

func TestGetAccount_ValidationAndRepoErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAccountRepositoryPort(ctrl)
	svc := NewAccountService(repo)

	_, err := svc.GetAccount(0)
	assert.ErrorIs(t, err, model.ErrAccountIDMustBePositive)

	repo.EXPECT().GetAccountBalance(nil, int64(1)).Return(decimal.Zero, model.ErrAccountNotFound)
	_, err = svc.GetAccount(1)
	assert.ErrorIs(t, err, model.ErrAccountNotFound)
}

func TestTransfer_ValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAccountRepositoryPort(ctrl)
	svc := NewAccountService(repo)

	err := svc.Transfer(0, 2, decimal.NewFromFloat(10))
	assert.ErrorIs(t, err, model.ErrAccountIDMustBePositive)

	err = svc.Transfer(1, 1, decimal.NewFromFloat(10))
	assert.ErrorIs(t, err, model.ErrSourceAndDestinationMustDiffer)

	err = svc.Transfer(1, 2, decimal.Zero)
	assert.ErrorIs(t, err, model.ErrAmountMustBePositive)
}

func TestTransfer_RepositoryAndTransactionErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAccountRepositoryPort(ctrl)
	tx := mocks.NewMockTransactionPort(ctrl)
	svc := NewAccountService(repo)

	sourceID, destID := int64(1), int64(2)
	amount := decimal.NewFromFloat(10)

	// BeginTx error
	repo.EXPECT().BeginTx().Return(nil, errors.New("begin tx error"))
	err := svc.Transfer(sourceID, destID, amount)
	assert.ErrorContains(t, err, "begin tx error")

	// Source account not found
	repo.EXPECT().BeginTx().Return(tx, nil)
	repo.EXPECT().GetAccountBalance(tx, sourceID).Return(decimal.Zero, model.ErrAccountNotFound)
	tx.EXPECT().Rollback()
	err = svc.Transfer(sourceID, destID, amount)
	assert.ErrorIs(t, err, model.ErrSourceAccountNotFound)
}

func TestTransfer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockAccountRepositoryPort(ctrl)
	tx := mocks.NewMockTransactionPort(ctrl)
	svc := NewAccountService(repo)

	sourceID, destID := int64(1), int64(2)
	amount := decimal.NewFromFloat(10)

	repo.EXPECT().BeginTx().Return(tx, nil)
	repo.EXPECT().GetAccountBalance(tx, sourceID).Return(decimal.NewFromFloat(20), nil)
	repo.EXPECT().GetAccountBalance(tx, destID).Return(decimal.NewFromFloat(5), nil)
	repo.EXPECT().UpdateAccountBalance(tx, sourceID, amount.Neg()).Return(nil)
	repo.EXPECT().UpdateAccountBalance(tx, destID, amount).Return(nil)
	tx.EXPECT().Commit().Return(nil)

	err := svc.Transfer(sourceID, destID, amount)
	assert.NoError(t, err)
}
