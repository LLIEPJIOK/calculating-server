package database

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	"github.com/LLIEPJIOK/calculating-server/internal/user"
)

func expressionSliceToString(expressions []*expression.Expression) string {
	result := "[ "
	for _, exp := range expressions {
		result += fmt.Sprint(*exp)
	}
	result += " ]"
	return result
}

func expContainsInSlice(expressions []*expression.Expression, target *expression.Expression) bool {
	for _, exp := range expressions {
		if exp.Equals(target) {
			return true
		}
	}
	return false
}

func checkMaxId(t *testing.T, login string, expectedMaxId uint64) {
	maxId, err := getMaxExpressionId(login)
	if err != nil {
		t.Fatalf("getting max id (error): expected %v, got: %v", nil, err)
	}
	if maxId != expectedMaxId {
		t.Fatalf("getting max id (maxId): expected %v, got: %v", expectedMaxId, maxId)
	}
}

func checkUncalculatingExpressions(t *testing.T, expressions []*expression.Expression) {
	uncalculatingExpressions := GetUncalculatingExpressions()
	for _, exp := range expressions {
		if (exp.Status == "calculating" || exp.Status == "in queue") && !expContainsInSlice(uncalculatingExpressions, exp) {
			t.Fatalf("expression %v isn't contained in uncalculating expressions %v", exp, expressionSliceToString(uncalculatingExpressions))
		}
	}
}

func checkLastExpressions(t *testing.T, expressions []*expression.Expression, login string) {
	lastExpressions := GetLastExpressions(login)
	counter := 0
	for i := len(expressions) - 1; i >= 0; i-- {
		if expressions[i].Login == login {
			if !expContainsInSlice(lastExpressions, expressions[i]) {
				t.Fatalf("expression %v isn't contained in last expressions %v", expressions[i], expressionSliceToString(lastExpressions))
			}
			counter++
			if counter == 10 {
				break
			}
		}
	}
	if len(lastExpressions) != counter {
		t.Fatalf("some extra expressions are contained in last expressions %v", expressionSliceToString(lastExpressions))
	}
}

func checkExpressionsById(t *testing.T, expressions []*expression.Expression) {
	for _, exp := range expressions {
		expressionFromDatabase := GetExpressionById(exp.Id, exp.Login)
		if !exp.Equals(expressionFromDatabase) {
			t.Fatalf("expression: expected: %v, but got: %v", exp, expressionFromDatabase)
		}
	}
}

func checkingExpressionsByExpression(t *testing.T, expressions []*expression.Expression, expressionString, userLogin string) {
	findExpressions := GetExpressionsByExpression(expressionString, userLogin)
	for _, exp := range expressions {
		if strings.Contains(exp.Exp, expressionString) && exp.Login == userLogin && !expContainsInSlice(findExpressions, exp) {
			t.Fatalf("expression %v isn't contained in find expressions %v", exp, expressionSliceToString(findExpressions))
		}
	}
}

func checkAll(t *testing.T, expressions []*expression.Expression) {
	t.Run("checking max id", func(t *testing.T) {
		checkMaxId(t, "1", 3)
		checkMaxId(t, "2", 11)
		checkMaxId(t, "3", 3)
	})

	t.Run("checking uncalculating expressions", func(t *testing.T) {
		checkUncalculatingExpressions(t, expressions)
	})

	t.Run("checking last expressions", func(t *testing.T) {
		checkLastExpressions(t, expressions, "1")
		checkLastExpressions(t, expressions, "2")
		checkLastExpressions(t, expressions, "3")
	})

	t.Run("checking expressions by id", func(t *testing.T) {
		checkExpressionsById(t, expressions)
	})

	t.Run("checking expressions by expression string", func(t *testing.T) {
		checkingExpressionsByExpression(t, expressions, "1+", "1")
		checkingExpressionsByExpression(t, expressions, "+", "3")
		checkingExpressionsByExpression(t, expressions, "/", "2")
	})
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
			CreationTime: time.Now().Add(-24 * time.Hour),
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
			Login:        "2",
			Exp:          "2 + 34 + 6 * 5 - 7",
			Result:       59,
			Status:       "calculating",
			Err:          "",
			CreationTime: time.Now().Add(-time.Second),
		},
		{
			Login:        "2",
			Exp:          "$$",
			Result:       0,
			Status:       "error",
			Err:          "some error",
			CreationTime: time.Now().Add(-time.Hour),
		},
		{
			Login:           "2",
			Exp:             "2 * 2 - 2 * 2 + 222",
			Result:          222,
			Status:          "calculated",
			Err:             "",
			CreationTime:    time.Now().Add(-time.Second),
			CalculationTime: time.Now().Add(-time.Millisecond),
		},
		{
			Login:           "2",
			Exp:             "(2 * 2 - 2 * 2 + 222)",
			Result:          222,
			Status:          "calculated",
			Err:             "",
			CreationTime:    time.Now().Add(-time.Microsecond),
			CalculationTime: time.Now().Add(-time.Nanosecond),
		},
		{
			Login:        "2",
			Exp:          "5 / (5 - 5)",
			Result:       0,
			Status:       "error",
			Err:          "some error",
			CreationTime: time.Now().Add(-time.Second),
		},
		{
			Login:           "2",
			Exp:             "1 + 1 - 1",
			Result:          1,
			Status:          "calculated",
			Err:             "",
			CreationTime:    time.Now().Add(-time.Second),
			CalculationTime: time.Now().Add(-time.Second),
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

	os.Setenv("expressionDatabaseName", "expression_table_test_db")
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

	t.Run("before update", func(t *testing.T) {
		checkAll(t, expressions)
	})

	expressions[0].Status = "calculated"
	expressions[0].Result = 2
	UpdateExpressionStatus(expressions[0])
	UpdateExpressionResult(expressions[0])

	expressions[3].Status = "calculated"
	UpdateExpressionStatus(expressions[3])

	expressions[15].Status = "calculating"
	UpdateExpressionStatus(expressions[15])

	t.Run("after update", func(t *testing.T) {
		checkAll(t, expressions)
	})
}
