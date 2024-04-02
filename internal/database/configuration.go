package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host             = "localhost"
	port             = 5432
	databaseUser     = "postgres"
	password         = "123409874567"
	ExpressionDBName = "expressions_db"
	OperationTimeDB  = "operation_time_db"
)

var (
	dataBase *sql.DB
)

func createDatabaseIfNotExists() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, databaseUser, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error open database:", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatal("error connecting to database:", err)
	}

	rows, err := db.Query(`
		SELECT 1 
		FROM pg_database 
		WHERE datname = $1`,
		ExpressionDBName)
	if err != nil {
		log.Fatal("error checking database existence:", err)
	}
	defer rows.Close()

	if !rows.Next() {
		_, err = db.Exec(`CREATE DATABASE ` + ExpressionDBName)
		if err != nil {
			log.Fatal("error creating database:", err)
		}
	}
}

func createTablesIfNotExists() {
	createExpressionsTableIfNotExists()
	createOperationTimeTableIfNotExists()
	createUsersTableIfNotExists()
}

func Close() {
	dataBase.Close()
}

func Configure() {
	createDatabaseIfNotExists()

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, databaseUser, password, ExpressionDBName)
	var err error
	dataBase, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("error open database:", err)
	}
	if err = dataBase.Ping(); err != nil {
		log.Fatal("error connecting to database:", err)
	}

	createTablesIfNotExists()
}
