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
		fmt.Printf(".env not found %v\n", err)
	}

	// Get database connection details from environment variables
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	if len(dbUser) == 0 {
		panic("dbUser not found")
	}
	if len(dbName) == 0 {
		panic("dbName not found")
	}

	// Create connection string
	connStr := fmt.Sprintf("postgres://%s:%s@localhost:5433/%s?sslmode=disable", dbUser, dbPassword, dbName)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test the connection
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
