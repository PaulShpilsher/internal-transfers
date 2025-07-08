package db

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	cleanup := func() { db.Close() }
	return db, mock, cleanup
}

func TestCreateAccountAndGetBalance(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)
	accountID := int64(1)
	initialBalance := decimal.NewFromInt(1000)

	// Expect insert
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO accounts (account_id, balance) VALUES ($1, $2)")).
		WithArgs(accountID, initialBalance.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.CreateAccount(accountID, initialBalance)
	assert.NoError(t, err)

	// Expect select
	rows := sqlmock.NewRows([]string{"balance"}).AddRow(initialBalance.String())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT balance FROM accounts WHERE account_id = $1 LIMIT 1")).
		WithArgs(accountID).
		WillReturnRows(rows)

	balance, err := repo.GetAccountBalance(nil, accountID)
	assert.NoError(t, err)
	assert.True(t, balance.Equal(initialBalance), "expected balance to match initial")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAccountBalance(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)
	accountID := int64(2)
	initialBalance := decimal.NewFromInt(500)

	// Expect insert
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO accounts (account_id, balance) VALUES ($1, $2)")).
		WithArgs(accountID, initialBalance.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	_ = repo.CreateAccount(accountID, initialBalance)

	// Expect begin
	mock.ExpectBegin()
	tx, err := repo.BeginTx()
	assert.NoError(t, err)

	delta := decimal.NewFromInt(200)
	// Expect update
	mock.ExpectExec(regexp.QuoteMeta("UPDATE accounts SET balance = balance + $1 WHERE account_id = $2")).
		WithArgs(delta.String(), accountID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateAccountBalance(tx, accountID, delta)
	assert.NoError(t, err)

	// Expect commit
	mock.ExpectCommit()
	err = tx.Commit()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Error case: CreateAccount returns error on DB failure
func TestCreateAccount_DBError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)
	accountID := int64(10)
	initialBalance := decimal.NewFromInt(100)

	dberr := sql.ErrConnDone
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO accounts (account_id, balance) VALUES ($1, $2)")).
		WithArgs(accountID, initialBalance.String()).
		WillReturnError(dberr)

	err := repo.CreateAccount(accountID, initialBalance)
	assert.Error(t, err)
	assert.Equal(t, dberr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Error case: GetAccountBalance returns ErrAccountNotFound on no rows
func TestGetAccountBalance_AccountNotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)
	accountID := int64(404)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT balance FROM accounts WHERE account_id = $1 LIMIT 1")).
		WithArgs(accountID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetAccountBalance(nil, accountID)
	assert.Error(t, err)
	assert.Equal(t, "account not found", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Error case: UpdateAccountBalance returns error on update failure
func TestUpdateAccountBalance_DBError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)
	accountID := int64(3)
	initialBalance := decimal.NewFromInt(200)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO accounts (account_id, balance) VALUES ($1, $2)")).
		WithArgs(accountID, initialBalance.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	_ = repo.CreateAccount(accountID, initialBalance)

	mock.ExpectBegin()
	tx, err := repo.BeginTx()
	assert.NoError(t, err)

	dberr := sql.ErrTxDone
	delta := decimal.NewFromInt(50)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE accounts SET balance = balance + $1 WHERE account_id = $2")).
		WithArgs(delta.String(), accountID).
		WillReturnError(dberr)

	err = repo.UpdateAccountBalance(tx, accountID, delta)
	assert.Error(t, err)
	assert.Equal(t, dberr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Edge case: GetAccountBalance returns error if decimal parsing fails
func TestGetAccountBalance_DecimalParseError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)
	accountID := int64(123)

	// Return a non-numeric string for balance
	rows := sqlmock.NewRows([]string{"balance"}).AddRow("not-a-number")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT balance FROM accounts WHERE account_id = $1 LIMIT 1")).
		WithArgs(accountID).
		WillReturnRows(rows)

	_, err := repo.GetAccountBalance(nil, accountID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not-a-number")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Edge case: UpdateAccountBalance returns error if tx is nil
func TestUpdateAccountBalance_NilTx(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)
	accountID := int64(5)
	delta := decimal.NewFromInt(10)

	err := repo.UpdateAccountBalance(nil, accountID, delta)
	assert.Error(t, err)
}

// Edge case: BeginTx returns error if DB cannot begin transaction
func TestBeginTx_DBError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()
	repo := NewAccountRepository(db)

	dberr := sql.ErrConnDone
	mock.ExpectBegin().WillReturnError(dberr)

	tx, err := repo.BeginTx()
	assert.Nil(t, tx)
	assert.Error(t, err)
	assert.Equal(t, dberr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
