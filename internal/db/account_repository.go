package db

import (
	"database/sql"
	"fmt"

	"internal-transfers/internal/model"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type AccountRepository struct {
	conn *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{conn: db}
}

// BeginTx starts a new transaction and returns the abstraction
func (p *AccountRepository) BeginTx() (*Transaction, error) {
	tx, err := p.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

func (p *AccountRepository) CreateAccount(accountID int64, initialBalance decimal.Decimal) error {
	_, err := p.conn.Exec(`INSERT INTO accounts (account_id, balance) VALUES ($1, $2)`, accountID, initialBalance.String)
	return err
}

func (p *AccountRepository) GetAccountBalance(tx *Transaction, accountID int64) (decimal.Decimal, error) {
	var balanceStr string
	var err error

	if tx != nil {
		err = tx.tx.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE LIMIT 1`, accountID).Scan(&balanceStr)
	} else {
		err = p.conn.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1 LIMIT 1`, accountID).Scan(&balanceStr)
	}

	if err == sql.ErrNoRows {
		return decimal.Zero, model.ErrAccountNotFound
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("query account by id: %w", err)
	}

	balance, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

// UpdateAccountBalanceTx updates the balance for an account within a transaction
func (p *AccountRepository) UpdateAccountBalance(tx *Transaction, accountID int64, delta decimal.Decimal) error {
	_, err := tx.tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`, delta.String(), accountID)
	return err
}

// func (t *PostgresTransaction) UpdateAccountBalance(accountID int64, delta decimal.Decimal) error {
// 	_, err := t.tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`, delta, accountID)
// 	return err
// }
// func (t *PostgresTransaction) InsertTransaction(sourceID, destID int64, amount decimal.Decimal) error {
// 	_, err := t.tx.Exec(`INSERT INTO transactions (source_account_id, destination_account_id, amount) VALUES ($1, $2, $3)`, sourceID, destID, amount)
// 	return err
// }
