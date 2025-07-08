package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type AccountRepository struct {
	Conn *sql.DB
}

func NewAccountRepository(dataSource string) (*AccountRepository, error) {
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &AccountRepository{Conn: db}, nil
}

func (p *AccountRepository) CreateAccount(accountID int64, initialBalance decimal.Decimal) error {
	_, err := p.Conn.Exec(`INSERT INTO accounts (account_id, balance) VALUES ($1, $2)`, accountID, initialBalance)
	return err
}

func (p *AccountRepository) GetAccount(accountID int64) (decimal.Decimal, error) {
	var balanceStr string
	err := p.Conn.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1`, accountID).Scan(&balanceStr)
	if err == sql.ErrNoRows {
		return decimal.Zero, errors.New("account not found")
	}
	if err != nil {
		return decimal.Zero, err
	}
	balance, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return decimal.Zero, err
	}
	return balance, nil
}

func (p *AccountRepository) Transfer(sourceID, destID int64, amount decimal.Decimal) error {
	tx, err := p.Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var sourceBalanceStr string
	err = tx.QueryRow(`SELECT balance FROM accounts WHERE account_id = $1 FOR UPDATE`, sourceID).Scan(&sourceBalanceStr)
	if err != nil {
		return fmt.Errorf("source account not found")
	}
	sourceBalance, err := decimal.NewFromString(sourceBalanceStr)
	if err != nil {
		return err
	}
	if sourceBalance.LessThan(amount) {
		return fmt.Errorf("insufficient funds")
	}

	_, err = tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE account_id = $2`, amount, sourceID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE account_id = $2`, amount, destID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO transactions (source_account_id, destination_account_id, amount) VALUES ($1, $2, $3)`, sourceID, destID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}
