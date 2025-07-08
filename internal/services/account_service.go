package services

import (
	"errors"
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

	// Check if account already exists
	_, err := s.Repo.GetAccount(account.AccountID)
	if err == nil {
		return errors.New("account id already exists")
	}
	if err != nil && err.Error() != "account not found" {
		return err // DB error
	}

	// Validate balance is nonnegative
	if account.Balance.IsNegative() {
		return errors.New("balance must be non-negative")
	}

	// Validate balance precision (scale)
	if account.Balance.Exponent() < -8 {
		return errors.New("balance precision must be 8 or fewer decimal places")
	}

	return s.Repo.CreateAccount(account.AccountID, account.Balance)
}
