package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase() (*Database, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic(".env file not found")
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	// Create connection string
	connStr := fmt.Sprintf("postgres://%s:%s@localhost:5433/%s?sslmode=disable", dbUser, dbPassword, dbName)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	db.SetConnMaxIdleTime(10 * time.Second)
	db.SetConnMaxLifetime(30 * time.Second)
	db.SetMaxIdleConns(0)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
