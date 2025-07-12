package db

import (
	"database/sql"
	"fmt"
	"os"

	// registering the mysql driver.
	_ "github.com/go-sql-driver/mysql"
)

// Connect establishes a connection to the MySQL database.
// It returns a pointer to a sql.DB object and an error if the connection fails.
func Connect() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPswd := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// set default if env are not set
	if dbUser == "" {
		dbUser = "root"
	}

	if dbPswd == "" {
		return nil, fmt.Errorf("DB_PASSWORD is not set")
	}

	if dbName == "" {
		dbHost = "localhost"
	}

	if dbPort == "" {
		dbPort = "3306"
	}

	// dsn (Data Source Name) string contains the connection parameters.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPswd, dbHost, dbPort, dbName)

	// sql.Open initializes a database handle. It does not create a connection.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// db.Ping verifies that a connection to the database is still alive,
	// establishing a connection if necessary.
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
