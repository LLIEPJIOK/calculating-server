package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var (
	dataBase *sql.DB
)

func createDatabaseIfNotExists() {
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("host"), os.Getenv("port"), os.Getenv("databaseUser"), os.Getenv("password"))
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
	)`, os.Getenv("expressionDatabaseName")).Scan(&exists)
	if err != nil {
		log.Fatal("error checking database existence:", err)
	}

	if !exists {
		_, err = dataBase.Exec(`CREATE DATABASE ` + os.Getenv("expressionDatabaseName"))
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

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("host"), os.Getenv("port"), os.Getenv("databaseUser"), os.Getenv("password"), os.Getenv("expressionDatabaseName"))
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
