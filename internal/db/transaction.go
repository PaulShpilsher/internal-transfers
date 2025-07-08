package db

import "database/sql"

//go:generate mockgen -destination=../mocks/mock_transaction_port.go -package=mocks internal-transfers/internal/db TransactionPort
type TransactionPort interface {
	Commit() error
	Rollback() error
}

type Transaction struct {
	tx *sql.Tx
}

func (t *Transaction) Commit() error   { return t.tx.Commit() }
func (t *Transaction) Rollback() error { return t.tx.Rollback() }
