package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host         = "localhost"
	port         = 5432
	databaseUser = "postgres"
	password     = "123409874567"
)

var (
	expressionDatabaseName = "expressions_db"

	dataBase *sql.DB
)

func createDatabaseIfNotExists() {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, databaseUser, password)
	dataBase, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("error open postgres:", err)
	}
	defer dataBase.Close()
	if err = dataBase.Ping(); err != nil {
		log.Fatal("error connecting to postgres:", err)
	}

	var exists bool
	err = dataBase.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_database 
			WHERE datname = $1
	)`, expressionDatabaseName).Scan(&exists)
	if err != nil {
		log.Fatal("error checking database existence:", err)
	}

	if !exists {
		_, err = dataBase.Exec(`CREATE DATABASE ` + expressionDatabaseName)
		if err != nil {
			log.Fatal("error creating database:", err)
		}
	}
}

func createTablesIfNotExists() {
	createUsersTableIfNotExists()
	createExpressionTableIfNotExists()
	createOperationsTimeTableIfNotExists()
}

func Close() {
	if dataBase != nil {
		if err := dataBase.Close(); err != nil {
			log.Printf("error while closing database: %v\n", err)
		}
	}
}

func Configure() {
	createDatabaseIfNotExists()

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, databaseUser, password, expressionDatabaseName)
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
