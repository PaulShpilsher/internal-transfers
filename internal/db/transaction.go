package db

import "database/sql"

// TransactionPort defines the interface for transaction management
//
//go:generate mockgen -destination=../mocks/mock_transaction_port.go -package=mocks internal-transfers/internal/db TransactionPort
type TransactionPort interface {
	Commit() error
	Rollback() error
}

// Transaction implements the TransactionPort interface for managing database transactions
type Transaction struct {
	tx *sql.Tx
}

func (t *Transaction) Commit() error   { return t.tx.Commit() }
func (t *Transaction) Rollback() error { return t.tx.Rollback() }
