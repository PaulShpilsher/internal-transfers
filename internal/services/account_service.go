package services

import (
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

	return s.Repo.CreateAccount(account.AccountID, account.Balance)
}
