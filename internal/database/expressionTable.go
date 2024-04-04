package database

import (
	"database/sql"
	"log"

	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	_ "github.com/lib/pq"
)

func rowsToExpressionsSlice(rows *sql.Rows) []*expression.Expression {
	var expressions []*expression.Expression
	for rows.Next() {
		var exp expression.Expression
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

func InsertExpression(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		INSERT INTO "expressions"(exp, result, status, creation_time, calculation_time) 
		VALUES($1, $2, $3, $4, $5)`,
		exp.Exp, exp.Result, exp.Status, exp.CreationTime, exp.CalculationTime)
	if err != nil {
		log.Printf("error in insert %#v in database: %v\n", *exp, err)
		return
	}

	err = dataBase.QueryRow(`SELECT LASTVAL()`).Scan(&exp.Id)
	if err != nil {
		log.Fatal(err)
	}
}

func GetExpressionById(id string) []*expression.Expression {
	var rows *sql.Rows
	var err error
	if id == "" {
		rows, err = dataBase.Query(`
			SELECT * 
			FROM "expressions" 
			ORDER BY id DESC`)
	} else {
		rows, err = dataBase.Query(`
			SELECT * 
			FROM "expressions" 
			WHERE CAST(id AS TEXT) LIKE '%' || $1 || '%'
			ORDER BY id ASC`, id)
	}
	if err != nil {
		log.Fatal("error in getting data from database:", err)
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func GetUncalculatingExpressions() []*expression.Expression {
	rows, err := dataBase.Query(`
		SELECT * 
		FROM "expressions"
		WHERE status = 'calculating' OR status = 'in queue'`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func UpdateExpressionStatus(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		UPDATE "expressions" 
		SET status = $1 
		WHERE id = $2`,
		exp.Status, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}
}

func UpdateExpressionResult(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		UPDATE "expressions" 
		SET result = $1 
		WHERE id = $2`,
		exp.Result, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}

	_, err = dataBase.Exec(`
		UPDATE "expressions" 
		SET calculation_time = $1 
		WHERE id = $2`,
		exp.CalculationTime, exp.Id)
	if err != nil {
		log.Printf("error in setting %#v in database: %v\n", *exp, err)
	}
}

func GetLastExpressions() []*expression.Expression {
	rows, err := dataBase.Query("SELECT * FROM \"expressions\" ORDER BY creation_time DESC LIMIT 10")
	if err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	defer rows.Close()

	lastInputs := make([]*expression.Expression, 0, 10)
	for rows.Next() {
		var exp expression.Expression
		err := rows.Scan(&exp.Id, &exp.Exp, &exp.Result, &exp.Status, &exp.CreationTime, &exp.CalculationTime)
		if err != nil {
			log.Println("error in getting data from database:", err)
			return nil
		}
		lastInputs = append(lastInputs, &exp)
	}

	if err := rows.Err(); err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	return lastInputs
}

func createExpressionsTableIfNotExists() {
	_, err := dataBase.Exec(`
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
}
