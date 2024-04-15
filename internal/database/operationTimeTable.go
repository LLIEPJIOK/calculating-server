package database

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func GetOperationsTime(userLogin string) (map[string]uint64, error) {
	rows, err := dataBase.Query(`
		SELECT
			operation, time
		FROM "Operations_Time"
		WHERE user_login = $1
		`, userLogin)
	if err != nil {
		return nil, fmt.Errorf("error getting data from database: %v", err)
	}

	operationTimes := make(map[string]uint64)
	for rows.Next() {
		var operation string
		var time uint64
		err := rows.Scan(&operation, &time)
		if err != nil {
			return nil, fmt.Errorf("error getting operations time: %v", err)
		}
		operationTimes[operation] = time
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting operations time: %v", err)
	}
	return operationTimes, nil
}

func InsertDefaultOperationTimes(userLogin string) {
	_, err := dataBase.Exec(`
		INSERT INTO "Operations_Time"(user_login, operation, time)
		VALUES 
			($1, 'time-plus', 200),
			($1, 'time-minus', 200),
			($1, 'time-multiply', 200),
			($1, 'time-divide', 200);
		`, userLogin)
	if err != nil {
		log.Fatal("error inserting default operations time:", err)
	}
}

func UpdateOperationTimes(timePlus, timeMinus, timeMultiply, timeDivide uint64, userLogin string) {
	operationTimes := map[string]uint64{
		"time-plus":     timePlus,
		"time-minus":    timeMinus,
		"time-multiply": timeMultiply,
		"time-divide":   timeDivide,
	}
	for operation, time := range operationTimes {
		_, err := dataBase.Exec(`
			UPDATE "Operations_Time"
			SET time = $1
			WHERE operation = $2 and user_login = $3
			`, time, operation, userLogin)
		if err != nil {
			log.Fatal("error updating operation:", err)
		}
	}
}

func createOperationsTimeTableIfNotExists() {
	_, err := dataBase.Exec(`
		CREATE TABLE IF NOT EXISTS "Operations_Time" (
			user_login TEXT REFERENCES "User"(login),
			operation TEXT,
			time INT NOT NULL,
			PRIMARY KEY (user_login, operation)
		)`)
	if err != nil {
		log.Fatal("error creating Operations_Time table:", err)
	}
}
