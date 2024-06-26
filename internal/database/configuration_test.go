package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("No .env file found")
	}
}

func checkDatabaseExistence() (bool, error) {
	var exists bool
	err := dataBase.QueryRow(`
		SELECT EXISTS (
			SELECT 1 
			FROM pg_database 
			WHERE datname = $1
	)`, os.Getenv("DATABASE_NAME")).Scan(&exists)
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
	Close()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
	db, err := sql.Open("postgres", connStr)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error while closing database: %v\n", err)
		}
	}()
	if err != nil {
		log.Fatal("error open postgres:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("error connecting to postgres:", err)
	}

	_ = db.QueryRow(`
		DELETE 
			FROM pg_database 
			WHERE datname = $1
		`, os.Getenv("DATABASE_NAME"))
}

func TestConfiguration(t *testing.T) {
	os.Setenv("DATABASE_NAME", "configuration_test_db")
	defer deleteDatabase()
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

		exists, err = checkTableExistence("Operations_Time")
		if err != nil {
			t.Fatalf("%v) error while checking OperationsTime table existence: %v", i, err)
		} else if !exists {
			t.Fatalf("%v) OperationsTime table hasn't been created", i)
		}
	}
}
