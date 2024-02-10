package expressions

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host             = "localhost"
	port             = 5432
	user             = "postgres"
	password         = "123409874567"
	ExpressionDBName = "expressions_db"
	OperationTimeDB  = "operation_time_db"
)

type Database struct {
	*sql.DB
	LastInputs []*Expression
}

func rowsToExpressionsSlice(rows *sql.Rows) []*Expression {
	var expressions []*Expression
	for rows.Next() {
		var exp Expression
		err := rows.Scan(&exp.Id, &exp.Exp, &exp.Result, &exp.Status, &exp.CreationTime, &exp.CalculationTime)
		if err != nil {
			log.Println("error in getting data from database:", err)
			return nil
		}
		expressions = append(expressions, &exp)
	}

	if err := rows.Err(); err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	return expressions
}

func createDB() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, user, password)
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

func (db *Database) createTables() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS "expressions" (
			id SERIAL PRIMARY KEY,
			exp CHARACTER VARYING,
			result DOUBLE PRECISION,
			status CHARACTER VARYING,
			creation_time TIMESTAMP,
			calculation_time TIMESTAMP
		)`)
	if err != nil {
		log.Fatal("error creating expressions table:", err)
	}

	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_name = 'operations_time'
		)`).Scan(&exists)
	if err != nil {
		log.Fatal("error getting information from database:", err)
	}

	if !exists {
		_, err = db.Exec(`
		CREATE TABLE "operations_time" (
			key CHARACTER VARYING PRIMARY KEY,
			value INT
		)`)
		if err != nil {
			log.Fatal("error creating expressions table:", err)
		}
	}

	rows, err := db.Query(`
		SELECT * 
		FROM "operations_time"`)
	if err != nil {
		log.Fatal("error getting data from db:", err)
	}

	for rows.Next() {
		var key string
		var val int64
		err := rows.Scan(&key, &val)
		if err != nil {
			log.Fatal("error in getting data from database:", err)
		}
		operationsTime[key] = val
	}
	if err := rows.Err(); err != nil {
		log.Fatal("error in getting data from database:", err)
	}

	for key, val := range operationsTime {
		var exists bool
		err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM "operations_time"
			WHERE key = $1
		)`, key).Scan(&exists)
		if err != nil {
			log.Fatal("error checking record existence:", err)
		}

		if !exists {
			_, err := db.Exec(`
			INSERT INTO "operations_time"(key, value)
			VALUES ($1, $2)`,
				key, val)
			if err != nil {
				log.Fatal("error inserting record:", err)
			}
		}
	}
}

func (db *Database) InsertExpressionInBD(exp *Expression) {
	_, err := db.Exec(`
		INSERT INTO "expressions"(exp, result, status, creation_time, calculation_time) 
		VALUES($1, $2, $3, $4, $5)`,
		exp.Exp, exp.Result, exp.Status, exp.CreationTime, exp.CalculationTime)
	if err != nil {
		log.Printf("error in insert %#v in database: %v\n", *exp, err)
		return
	}

	err = db.QueryRow(`SELECT LASTVAL()`).Scan(&exp.Id)
	if err != nil {
		log.Fatal(err)
	}

	db.LastInputs = append([]*Expression{exp}, db.LastInputs...)
	if len(db.LastInputs) >= 11 {
		db.LastInputs = db.LastInputs[:10]
	}
}

func (db *Database) GetExpressionById(id string) []*Expression {
	rows, err := db.Query(`
		SELECT * 
		FROM "expressions" 
		WHERE CAST(id AS TEXT) LIKE '%' || $1 || '%'
		ORDER BY creation_time DESC`, id)
	if err != nil {
		log.Fatal("error in getting data from database:", err)
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func (*Database) GetOperationTime(op string) int64 {
	return operationsTime[op]
}

func (db *Database) GetUncalculatingExpressions() []*Expression {
	rows, err := db.Query(`
		SELECT * 
		FROM "expressions"
		WHERE status = 'calculating' OR status = 'in queue'`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func (db *Database) UpdateStatus(exp *Expression) {
	_, err := db.Exec(`
		UPDATE "expressions" 
		SET status = $1 
		WHERE id = $2`,
		exp.Status, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}
}

func (db *Database) UpdateResult(exp *Expression) {
	_, err := db.Exec(`
		UPDATE "expressions" 
		SET result = $1 
		WHERE id = $2`,
		exp.Result, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}

	_, err = db.Exec(`
		UPDATE "expressions" 
		SET calculation_time = $1 
		WHERE id = $2`,
		exp.CalculationTime, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}
}

func (db *Database) UpdateOperationsTime(timePlus, timeMinus, timeMultiply, timeDivide int64) {
	operationsTime["time_plus"] = timePlus
	operationsTime["time_minus"] = timeMinus
	operationsTime["time_multiply"] = timeMultiply
	operationsTime["time_divide"] = timeDivide

	for key, val := range operationsTime {
		_, err := db.Exec(`
			UPDATE "operations_time"
			SET value = $2
			WHERE key = $1`,
			key, val)
		if err != nil {
			log.Fatal("error updating operation:", err)
		}
	}
}

func (db *Database) LoadLastExpressions(count int) {
	rows, err := db.Query("SELECT * FROM \"expressions\" ORDER BY creation_time DESC LIMIT $1", count)
	if err != nil {
		log.Fatal("error in getting data from database:", err)
	}
	defer rows.Close()

	db.LastInputs = make([]*Expression, 0)
	for rows.Next() {
		var exp Expression
		err := rows.Scan(&exp.Id, &exp.Exp, &exp.Result, &exp.Status, &exp.CreationTime, &exp.CalculationTime)
		if err != nil {
			log.Fatal("error in getting data from database:", err)
		}
		db.LastInputs = append(db.LastInputs, &exp)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("error in getting data from database:", err)
	}
}

func NewDB() *Database {
	db := &Database{}
	createDB()

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, ExpressionDBName)
	var err error
	db.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error open database:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("error connecting to database:", err)
	}

	db.createTables()
	db.LoadLastExpressions(10)
	return db
}
