package db

import (
	"database/sql"
	"fmt"
	"os"
)

// NewDBConnection creates and returns a new database connection using environment variables
func NewDBConnection() (*sql.DB, error) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	if err := dbConn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return dbConn, nil
}
