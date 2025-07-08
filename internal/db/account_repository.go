package db

import (
	"database/sql"
	"fmt"
	"log"

	"internal-transfers/internal/model"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// AccountRepositoryPort defines the repository interface for accounts
//
//go:generate mockgen -destination=../mocks/mock_account_repository.go -package=mocks internal-transfers/internal/db AccountRepositoryPort
type AccountRepositoryPort interface {
	BeginTx() (TransactionPort, error)
	CreateAccount(accountID int64, initialBalance decimal.Decimal) error
	GetAccountBalance(tx TransactionPort, accountID int64) (decimal.Decimal, error)
	UpdateAccountBalance(tx TransactionPort, accountID int64, delta decimal.Decimal) error
}

type AccountRepository struct {
	conn *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{conn: db}
}

// BeginTx starts a new transaction and returns the abstraction
func (repo *AccountRepository) BeginTx() (TransactionPort, error) {
	tx, err := repo.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

// CreateAccount creates a new account with the specified ID and initial balance
func (repo *AccountRepository) CreateAccount(accountID int64, initialBalance decimal.Decimal) error {
	_, err := repo.conn.Exec(`INSERT INTO accounts (account_id, balance) VALUES ($1, $2)`, accountID, initialBalance.String())
	if err != nil {
		log.Printf("CreateAccount DB error: %v", err)
	}
	return err
}

// GetAccountBalance retrieves the balance for an account, optionally within a transaction
func (repo *AccountRepository) GetAccountBalance(tx TransactionPort, accountID int64) (decimal.Decimal, error) {
	var balanceStr string
	var err error

	if tx != nil {
		dbTx, ok := tx.(*Transaction)
		if !ok {
			return decimal.Zero, fmt.Errorf("invalid transaction type")
		}
		err = dbTx.tx.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE LIMIT 1`, accountID).Scan(&balanceStr)
	} else {
		err = repo.conn.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1 LIMIT 1`, accountID).Scan(&balanceStr)
	}

	if err == sql.ErrNoRows {
		return decimal.Zero, model.ErrAccountNotFound
	}
	if err != nil {
		log.Printf("GetAccountBalance DB error: %v", err)
		return decimal.Zero, fmt.Errorf("query account by id: %w", err)
	}

	balance, err := decimal.NewFromString(balanceStr)
	if err != nil {
		log.Printf("GetAccountBalance parse error: %v", err)
		return decimal.Zero, err
	}
	return balance, nil
}

// UpdateAccountBalanceTx updates the balance for an account within a transaction
func (repo *AccountRepository) UpdateAccountBalance(tx TransactionPort, accountID int64, delta decimal.Decimal) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	dbTx, ok := tx.(*Transaction)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}
	_, err := dbTx.tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`, delta.String(), accountID)
	if err != nil {
		log.Printf("UpdateAccountBalanceTx DB error: %v", err)
	}
	return err
}
