package db

import "database/sql"

type Transaction struct {
	tx *sql.Tx
}

func (t *Transaction) Commit() error   { return t.tx.Commit() }
func (t *Transaction) Rollback() error { return t.tx.Rollback() }
