package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	_ "github.com/lib/pq"
)

func rowsToExpressionsSlice(rows *sql.Rows) []*expression.Expression {
	var expressions []*expression.Expression
	for rows.Next() {
		var exp expression.Expression
		err := rows.Scan(&exp.Login, &exp.Id, &exp.Exp, &exp.Result, &exp.Status, &exp.Err, &exp.CreationTime, &exp.CalculationTime)
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

func getMaxExpressionId(userLogin string) (uint64, error) {
	var maxId uint64
	err := dataBase.QueryRow(`
		SELECT 
			COALESCE(MAX(id), 0)
		FROM "Expression"
		WHERE user_login = $1
		`, userLogin).Scan(&maxId)
	if err != nil {
		return 0, fmt.Errorf("error getting max id where user_login = %v: %v", userLogin, err)
	}
	return maxId, nil
}

func InsertExpression(exp *expression.Expression) {
	prevId, err := getMaxExpressionId(exp.Login)
	if err != nil {
		log.Println(err)
		return
	}

	exp.Id = prevId + 1
	_, err = dataBase.Exec(`
		INSERT INTO "Expression"(user_login, id, exp, result, status, err, creation_time, calculation_time) 
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		`, exp.Login, exp.Id, exp.Exp, exp.Result, exp.Status, exp.Err, exp.CreationTime, exp.CalculationTime)
	if err != nil {
		log.Printf("error in insert %#v in database: %v\n", *exp, err)
		return
	}
}

func GetExpressionById(id uint64, userLogin string) *expression.Expression {
	exp := &expression.Expression{}
	err := dataBase.QueryRow(`
		SELECT 
			user_login, id, exp, result, status, err, creation_time, calculation_time
		FROM "Expression"
		WHERE id = $1 and user_login = $2
		ORDER BY creation_time DESC 
		`, id, userLogin).Scan(&exp.Login, &exp.Id, &exp.Exp, &exp.Result, &exp.Status, &exp.Err, &exp.CreationTime, &exp.CalculationTime)
	if err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	return exp
}

func GetExpressionsByExpression(exp, userLogin string) []*expression.Expression {
	rows, err := dataBase.Query(`
		SELECT 
			user_login, id, exp, result, status, err, creation_time, calculation_time
		FROM "Expression"
		WHERE exp LIKE '%' || $1 || '%' and user_login = $2
		ORDER BY creation_time DESC
		`, exp, userLogin)
	if err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func GetUncalculatingExpressions() []*expression.Expression {
	rows, err := dataBase.Query(`
		SELECT
			user_login, id, exp, result, status, err, creation_time, calculation_time
		FROM "Expression"
		WHERE status = 'calculating' OR status = 'in queue'
		`)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer rows.Close()

	expressions := rowsToExpressionsSlice(rows)
	for _, exp := range expressions {
		exp.OperationsTimes, err = GetOperationsTime(exp.Login)
		if err != nil {
			log.Println(err)
			return nil
		}
	}
	return expressions
}

func UpdateExpressionStatus(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		UPDATE "Expression"
		SET status = $1, err = $2
		WHERE user_login = $3 AND id = $4
		`, exp.Status, exp.Err, exp.Login, exp.Id)
	if err != nil {
		log.Printf("error in updating %#v in database: %v\n", *exp, err)
	}
}

func UpdateExpressionResult(exp *expression.Expression) {
	_, err := dataBase.Exec(`
		UPDATE "Expression"
		SET result = $1, calculation_time = $2
		WHERE user_login = $3 AND id = $4
		`, exp.Result, exp.CalculationTime, exp.Login, exp.Id)
	if err != nil {
		log.Printf("error in updating %#v in database: %v\n", *exp, err)
	}
}

func GetLastExpressions(userLogin string) []*expression.Expression {
	rows, err := dataBase.Query(`
		SELECT 
			user_login, id, exp, result, status, err, creation_time, calculation_time
		FROM "Expression"
		WHERE user_login = $1
		ORDER BY creation_time DESC 
		LIMIT 10
		`, userLogin)
	if err != nil {
		log.Println("error in getting data from database:", err)
		return nil
	}
	defer rows.Close()

	return rowsToExpressionsSlice(rows)
}

func createExpressionTableIfNotExists() {
	_, err := dataBase.Exec(`
		CREATE TABLE IF NOT EXISTS "Expression" (
			user_login TEXT REFERENCES "User"(login),
			id INT,
			exp TEXT NOT NULL,
			result NUMERIC NOT NULL,
			status TEXT NOT NULL,
			err TEXT NOT NULL,
			creation_time TIMESTAMP WITH TIME ZONE NOT NULL,
			calculation_time TIMESTAMP WITH TIME ZONE NOT NULL,
			PRIMARY KEY (user_login, id)
		)`)
	if err != nil {
		log.Fatal("error creating expression table:", err)
	}
}
