package services

import (
	"errors"
	"fmt"
	"internal-transfers/internal/db"
	"internal-transfers/internal/model"
	"log"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

const maxDecimalPrecision = 8

// AccountServicePort defines the service interface for accounts
//
//go:generate mockgen -destination=../mocks/mock_account_service.go -package=mocks internal-transfers/internal/services AccountServicePort
type AccountServicePort interface {
	CreateAccount(account model.Account) error
	GetAccount(id int64) (model.Account, error)
	Transfer(sourceID, destID int64, amount decimal.Decimal) error
}

type AccountService struct {
	repo db.AccountRepositoryPort
}

func NewAccountService(repo db.AccountRepositoryPort) *AccountService {
	return &AccountService{repo: repo}
}

// Validation helpers
func validateAccountID(id int64) error {
	if id <= 0 {
		return model.ErrAccountIDMustBePositive
	}
	return nil
}

func validateDecimalPrecision(val decimal.Decimal) error {
	if val.Exponent() < -maxDecimalPrecision {
		return model.ErrPrecisionTooHigh
	}
	return nil
}

func (s *AccountService) CreateAccount(account model.Account) error {
	if err := validateAccountID(account.AccountID); err != nil {
		log.Printf("CreateAccount validation failed: %v", err)
		return err
	}
	if account.Balance.IsNegative() {
		log.Printf("CreateAccount negative balance: %v", account.Balance)
		return model.ErrBalanceMustBeNonNegative
	}
	if err := validateDecimalPrecision(account.Balance); err != nil {
		log.Printf("CreateAccount balance precision error: %v", err)
		return err
	}

	err := s.repo.CreateAccount(account.AccountID, account.Balance)
	if err != nil {
		// Handle unique constraint violation (Postgres error code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Printf("CreateAccount duplicate account id: %d", account.AccountID)
			return model.ErrAccountIDAlreadyExists
		}
		log.Printf("CreateAccount db error: %v", err)
		return err
	}
	log.Printf("Account created: %d", account.AccountID)
	return nil
}

func (s *AccountService) GetAccount(id int64) (model.Account, error) {
	if err := validateAccountID(id); err != nil {
		log.Printf("GetAccount validation failed: %v", err)
		return model.Account{}, err
	}
	balance, err := s.repo.GetAccountBalance(nil, id)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			return model.Account{}, model.ErrAccountNotFound
		}
		log.Printf("GetAccount db error: %v", err)
		return model.Account{}, fmt.Errorf("get account: %w", err)
	}

	account := model.Account{
		AccountID: id,
		Balance:   balance,
	}
	return account, nil
}

func (s *AccountService) Transfer(sourceID, destID int64, amount decimal.Decimal) (err error) {
	if err = validateAccountID(sourceID); err != nil {
		log.Printf("Transfer validation failed for sourceID: %v", err)
		return err
	}
	if err = validateAccountID(destID); err != nil {
		log.Printf("Transfer validation failed for destID: %v", err)
		return err
	}
	if sourceID == destID {
		log.Printf("Transfer attempted with same source and destination: %d", sourceID)
		return model.ErrSourceAndDestinationMustDiffer
	}
	if amount.IsNegative() || amount.IsZero() {
		log.Printf("Transfer with non-positive amount: %v", amount)
		return model.ErrAmountMustBePositive
	}
	if err = validateDecimalPrecision(amount); err != nil {
		log.Printf("Transfer amount precision error: %v", err)
		return err
	}

	txn, err := s.repo.BeginTx()
	if err != nil {
		log.Printf("Transfer failed to begin transaction: %v", err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			txn.Rollback()
			log.Printf("Transfer panic, transaction rolled back: %v", p)
			panic(p)
		} else if err != nil {
			txn.Rollback()
			log.Printf("Transfer error, transaction rolled back: %v", err)
		}
	}()

	// Lock source account row and get balance
	balance, err := s.repo.GetAccountBalance(txn, sourceID)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			log.Printf("Transfer source account not found: %d", sourceID)
			return model.ErrSourceAccountNotFound
		}
		log.Printf("Transfer error getting source balance: %v", err)
		return err
	}
	if balance.LessThan(amount) {
		log.Printf("Transfer insufficient funds: %d, balance: %v, amount: %v", sourceID, balance, amount)
		return model.ErrInsufficientFunds
	}

	// Lock destination account row to ensure it exists
	_, err = s.repo.GetAccountBalance(txn, destID)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			log.Printf("Transfer destination account not found: %d", destID)
			return model.ErrDestinationAccountNotFound
		}
		log.Printf("Transfer error getting destination balance: %v", err)
		return err
	}

	// Update balances
	if err = s.repo.UpdateAccountBalance(txn, sourceID, amount.Neg()); err != nil {
		log.Printf("Transfer error updating source balance: %v", err)
		return err
	}
	if err = s.repo.UpdateAccountBalance(txn, destID, amount); err != nil {
		log.Printf("Transfer error updating destination balance: %v", err)
		return err
	}

	if err = txn.Commit(); err != nil {
		log.Printf("Transfer commit failed: %v", err)
		return err
	}
	log.Printf("Transfer successful: %d -> %d, amount: %v", sourceID, destID, amount)
	return nil
}
