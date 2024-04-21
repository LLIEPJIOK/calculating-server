package database

import (
	"os"
	"strconv"
	"testing"

	"github.com/LLIEPJIOK/calculating-server/internal/user"
)

type userOperationsTime struct {
	userLogin         string
	operationsTimeMap map[string]uint64
}

var (
	defaultOperationsTimeMap = map[string]uint64{
		"time-plus":     200,
		"time-minus":    200,
		"time-multiply": 200,
		"time-divide":   200,
	}
)

func compareOperationsTimesMap(first, second map[string]uint64) bool {
	if len(first) != len(second) {
		return false
	}

	for key, firstValue := range first {
		if secondValue, ok := second[key]; !ok || secondValue != firstValue {
			return false
		}
	}

	return true
}

func checkOperationsTime(t *testing.T, login string, operationsTimeMap map[string]uint64) {
	operationsTimeMapFromDatabase, err := GetOperationsTime(login)
	if err != nil {
		t.Fatalf("error: expected: %v, but got: %v", nil, err)
	}
	if !compareOperationsTimesMap(operationsTimeMap, operationsTimeMapFromDatabase) {
		t.Fatalf("operations map: expected: %v, but got: %v", operationsTimeMap, operationsTimeMapFromDatabase)
	}
}

func TestOperationTimeTable(t *testing.T) {
	userOperationsTimes := []userOperationsTime{
		{
			userLogin: "1",
			operationsTimeMap: map[string]uint64{
				"time-plus":     909,
				"time-minus":    90,
				"time-multiply": 9,
				"time-divide":   0,
			},
		},
		{
			userLogin: "2",
			operationsTimeMap: map[string]uint64{
				"time-plus":     4,
				"time-minus":    3,
				"time-multiply": 2,
				"time-divide":   1,
			},
		},
		{
			userLogin: "3",
			operationsTimeMap: map[string]uint64{
				"time-plus":     654,
				"time-minus":    343,
				"time-multiply": 65,
				"time-divide":   777,
			},
		},
	}

	os.Setenv("expressionDatabaseName", "operation_time_table_test_db")
	Configure()
	defer deleteDatabase()

	for i := 1; i <= 3; i++ {
		InsertUser(&user.User{
			Login: strconv.Itoa(i),
		})
	}

	for _, userOperationsTime := range userOperationsTimes {
		InsertDefaultOperationTimes(userOperationsTime.userLogin)
	}

	for _, userOperationsTime := range userOperationsTimes {
		checkOperationsTime(t, userOperationsTime.userLogin, defaultOperationsTimeMap)
	}

	for _, userOperationsTime := range userOperationsTimes {
		operationMap := userOperationsTime.operationsTimeMap
		UpdateOperationTimes(operationMap["time-plus"], operationMap["time-minus"], operationMap["time-multiply"],
			operationMap["time-divide"], userOperationsTime.userLogin)
	}

	for _, userOperationsTime := range userOperationsTimes {
		checkOperationsTime(t, userOperationsTime.userLogin, userOperationsTime.operationsTimeMap)
	}
}
