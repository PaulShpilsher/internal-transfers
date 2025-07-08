package services

import (
	"errors"
	"fmt"
	"internal-transfers/internal/model"

	"internal-transfers/internal/db"
)

type AccountService struct {
	Repo *db.AccountRepository
}

func NewAccountService(repo *db.AccountRepository) *AccountService {
	return &AccountService{Repo: repo}
}

func (s *AccountService) CreateAccount(account model.Account) error {
	// Validate account ID is nonnegative
	if account.AccountID < 0 {
		return errors.New("account id must be a nonnegative number")
	}

	// Validate balance is nonnegative
	if account.Balance.IsNegative() {
		return errors.New("balance must be non-negative")
	}

	// Validate balance precision (scale)
	if account.Balance.Exponent() < -8 {
		return errors.New("balance precision must be 8 or fewer decimal places")
	}

	// Check if account already exists
	_, err := s.Repo.GetAccountBalance(account.AccountID)
	if err == nil {
		return errors.New("account id already exists")
	} else if !errors.Is(err, model.ErrAccountNotFound) {
		return err
	}

	return s.Repo.CreateAccount(account.AccountID, account.Balance)
}

func (s *AccountService) GetAccount(id int64) (model.Account, error) {
	balance, err := s.Repo.GetAccountBalance(id)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			return model.Account{}, model.ErrAccountNotFound
		}
		return model.Account{}, fmt.Errorf("get account: %w", err)
	}

	account := model.Account{
		AccountID: id,
		Balance:   balance,
	}

	return account, nil
}
