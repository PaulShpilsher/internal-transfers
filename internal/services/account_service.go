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

type AccountService struct {
	Repo *db.AccountRepository
}

func NewAccountService(repo *db.AccountRepository) *AccountService {
	return &AccountService{Repo: repo}
}

// Validation helpers
func validateAccountID(id int64) error {
	if id <= 0 {
		return errors.New("account id must be a positive number")
	}
	return nil
}

func validateDecimalPrecision(val decimal.Decimal) error {
	if val.Exponent() < -maxDecimalPrecision {
		return fmt.Errorf("precision must be %d or fewer decimal places", maxDecimalPrecision)
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
		return errors.New("balance must be non-negative")
	}
	if err := validateDecimalPrecision(account.Balance); err != nil {
		log.Printf("CreateAccount balance precision error: %v", err)
		return err
	}

	err := s.Repo.CreateAccount(account.AccountID, account.Balance)
	if err != nil {
		// Handle unique constraint violation (Postgres error code 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			log.Printf("CreateAccount duplicate account id: %d", account.AccountID)
			return errors.New("account id already exists")
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
	balance, err := s.Repo.GetAccountBalance(nil, id)
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
		return errors.New("source and destination accounts must be different")
	}
	if amount.IsNegative() || amount.IsZero() {
		log.Printf("Transfer with non-positive amount: %v", amount)
		return errors.New("amount must be positive")
	}
	if err = validateDecimalPrecision(amount); err != nil {
		log.Printf("Transfer amount precision error: %v", err)
		return err
	}

	txn, err := s.Repo.BeginTx()
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
	balance, err := s.Repo.GetAccountBalance(txn, sourceID)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			log.Printf("Transfer source account not found: %d", sourceID)
			return errors.New("source account not found")
		}
		log.Printf("Transfer error getting source balance: %v", err)
		return err
	}
	if balance.LessThan(amount) {
		log.Printf("Transfer insufficient funds: %d, balance: %v, amount: %v", sourceID, balance, amount)
		return errors.New("insufficient funds")
	}

	// Lock destination account row to ensure it exists
	_, err = s.Repo.GetAccountBalance(txn, destID)
	if err != nil {
		if errors.Is(err, model.ErrAccountNotFound) {
			log.Printf("Transfer destination account not found: %d", destID)
			return errors.New("destination account not found")
		}
		log.Printf("Transfer error getting destination balance: %v", err)
		return err
	}

	// Update balances
	if err = s.Repo.UpdateAccountBalance(txn, sourceID, amount.Neg()); err != nil {
		log.Printf("Transfer error updating source balance: %v", err)
		return err
	}
	if err = s.Repo.UpdateAccountBalance(txn, destID, amount); err != nil {
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
