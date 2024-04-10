package database

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
)

func checkDatabaseExistence() (bool, error) {
	var exists bool
	err := dataBase.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_database 
			WHERE datname = $1
	)`, expressionDatabaseName).Scan(&exists)
	return exists, err
}

func checkTableExistence(tableName string) (bool, error) {
	var exists bool
	err := dataBase.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM information_schema.tables 
			WHERE table_name = $1
		)`, tableName).Scan(&exists)
	return exists, err
}

func deleteDatabase() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, databaseUser, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error open postgres:", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatal("error connecting to postgres:", err)
	}

	_ = db.QueryRow(`
		DROP DATABASE "$1"
		`, expressionDatabaseName)

}

func TestConfiguration(t *testing.T) {
	t.Parallel()
	expressionDatabaseName = "test_db"
	defer Close()
	for i := 0; i < 2; i++ {
		Configure()

		exists, err := checkDatabaseExistence()
		if err != nil {
			t.Fatalf("%v) error while checking database existence: %v", i, err)
		} else if !exists {
			t.Fatalf("%v) database hasn't been created", i)
		}

		exists, err = checkTableExistence("User")
		if err != nil {
			t.Fatalf("%v) error while checking User table existence: %v", i, err)
		} else if !exists {
			t.Fatalf("%v) User table hasn't been created", i)
		}

		exists, err = checkTableExistence("Expression")
		if err != nil {
			t.Fatalf("%v) error while checking Expression table existence: %v", i, err)
		} else if !exists {
			t.Fatalf("%v) Expression table hasn't been created", i)
		}

		exists, err = checkTableExistence("OperationsTime")
		if err != nil {
			t.Fatalf("%v) error while checking OperationsTime table existence: %v", i, err)
		} else if !exists {
			t.Fatalf("%v) OperationsTime table hasn't been created", i)
		}
	}

	deleteDatabase()
}
