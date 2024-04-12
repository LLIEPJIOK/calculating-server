package database

import (
	"strconv"
	"testing"
	"time"

	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	"github.com/LLIEPJIOK/calculating-server/internal/user"
)

func checkMaxId(t *testing.T, login string, expectedMaxId uint64) {
	maxId, err := getMaxExpressionId(login)
	if err != nil {
		t.Fatalf("getting max id (error): expected %v, got: %v", nil, err)
	}
	if maxId != expectedMaxId {
		t.Fatalf("getting max id (maxId): expected %v, got: %v", expectedMaxId, maxId)
	}
}

func ExpContainsInSlice(expressions []*expression.Expression, target *expression.Expression) bool {
	for _, exp := range expressions {
		if exp.Equals(target) {
			return true
		}
	}
	return false
}

func TestExpressionTable(t *testing.T) {
	expressions := []*expression.Expression{
		{
			Login:        "1",
			Exp:          "1+1",
			Result:       0,
			Status:       "calculating",
			Err:          "",
			CreationTime: time.Now().Add(-time.Hour),
		},
		{
			Login:           "1",
			Exp:             "1+2",
			Result:          3,
			Status:          "calculated",
			Err:             "",
			CreationTime:    time.Now().Add(-time.Minute),
			CalculationTime: time.Now(),
		},
		{
			Login:        "1",
			Exp:          "1++",
			Result:       0,
			Status:       "error",
			Err:          "some error",
			CreationTime: time.Now().Add(-time.Minute),
		},
		{
			Login:        "2",
			Exp:          "(1-1)",
			Result:       0,
			Status:       "in queue",
			Err:          "",
			CreationTime: time.Now().Add(-time.Second),
		},
		{
			Login:        "2",
			Exp:          "()",
			Result:       0,
			Status:       "error",
			Err:          "some error",
			CreationTime: time.Now().Add(-time.Second),
		},
		{
			Login:        "2",
			Exp:          "-",
			Result:       0,
			Status:       "error",
			Err:          "some error",
			CreationTime: time.Now().Add(-time.Second),
		},
		{
			Login:           "2",
			Exp:             "10 - 10 + 10 - 10 + 10",
			Result:          10,
			Status:          "calculated",
			Err:             "",
			CreationTime:    time.Now().Add(-time.Hour),
			CalculationTime: time.Now().Add(-time.Second),
		},
		{
			Login:        "2",
			Exp:          "100",
			Result:       100,
			Status:       "calculating",
			Err:          "",
			CreationTime: time.Now().Add(-time.Hour),
		},
		{
			Login:           "3",
			Exp:             "(2 + 2) * 2",
			Result:          8,
			Status:          "calculated",
			Err:             "",
			CreationTime:    time.Now().Add(-time.Second),
			CalculationTime: time.Now().Add(-time.Nanosecond),
		},
		{
			Login:        "3",
			Exp:          "(5 + 32 - 8 * 9 / 6) * 0",
			Result:       0,
			Status:       "in queue",
			Err:          "",
			CreationTime: time.Now().Add(-time.Nanosecond),
		},
		{
			Login:        "3",
			Exp:          "5 + 32 - 8 * 9 / 6) * 0",
			Result:       0,
			Status:       "error",
			Err:          "some error",
			CreationTime: time.Now().Add(-time.Minute),
		},
	}

	expressionDatabaseName = "expression_test_db"
	Configure()
	defer deleteDatabase()

	for i := 1; i <= 3; i++ {
		InsertUser(&user.User{
			Login: strconv.Itoa(i),
		})
	}

	for _, exp := range expressions {
		InsertExpression(exp)
	}

	checkMaxId(t, "1", 3)
	checkMaxId(t, "2", 5)
	checkMaxId(t, "3", 3)

	uncalculatingExpressions := GetUncalculatingExpressions()
	for _, exp := range expressions {
		if (exp.Status == "calculating" || exp.Status == "in queue") && ExpContainsInSlice(uncalculatingExpressions, exp) {
			t.Errorf("expression %#v isn't contained in uncalculating expressions %#v", exp, uncalculatingExpressions)
		}
	}
}
