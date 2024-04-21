package workers

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/expression"
	"github.com/LLIEPJIOK/calculating-server/internal/user"
	"github.com/joho/godotenv"
)

type expressionTestCase struct {
	exp            expression.Expression
	expectedStatus string
	expectedErr    string
	expectedResult float64
}

func init() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("No .env file found")
	}
}

func deleteDatabase() {
	database.Close()

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

func workersCalculating() bool {
	for i := range numberOfWorkers {
		if Workers.GetStatus(i) == "Calculation expression..." {
			return true
		}
	}
	return false
}

func TestWorkers(t *testing.T) {
	os.Setenv("DATABASE_NAME", "workers_test_db")
	database.Configure()
	defer deleteDatabase()

	testCases := []expressionTestCase{
		{
			exp:            expression.New("0", "2+2*2"),
			expectedStatus: "calculated",
			expectedErr:    "",
			expectedResult: 6,
		},
		{
			exp:            expression.New("0", "(2+2)*2"),
			expectedStatus: "calculated",
			expectedErr:    "",
			expectedResult: 8,
		},
		{
			exp:            expression.New("0", "1 + 5 / 5 * 4 - 34"),
			expectedStatus: "calculated",
			expectedErr:    "",
			expectedResult: -29,
		},
		{
			exp:            expression.New("1", "(1+1+1+1) / 2 * 3"),
			expectedStatus: "calculated",
			expectedErr:    "",
			expectedResult: 6,
		},
		{
			exp:            expression.New("0", "14 / 7 * 35 / 5 - 12"),
			expectedStatus: "calculated",
			expectedErr:    "",
			expectedResult: 2,
		},
		{
			exp:            expression.New("2", "q"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: unknown math symbol: q",
			expectedResult: 0,
		},
		{
			exp:            expression.New("1", "5 - 1 -"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of operations",
			expectedResult: 0,
		},
		{
			exp:            expression.New("1", "2345 // 6"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of operations",
			expectedResult: 0,
		},
		{
			exp:            expression.New("2", "14 / (5 * 4 - 20)"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: division by zero",
			expectedResult: 0,
		},
		{
			exp:            expression.New("0", "(45 - 0"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
			expectedResult: 0,
		},
	}
	for i := 0; i < 3; i++ {
		database.InsertUser(&user.User{
			Login: strconv.Itoa(i),
		})
	}
	for i := 0; i < 6; i++ {
		testCases[i].exp.Parse()
		if testCases[i].exp.Status != "error" {
			testCases[i].exp.Status = "in queue"
		}
		database.InsertExpression(&testCases[i].exp)
	}

	Initialize()
	defer CloseExpressionsChan()

	for i := 6; i < 10; i++ {
		testCases[i].exp.Parse()
		if testCases[i].exp.Status != "error" {
			testCases[i].exp.Status = "in queue"
		}
		database.InsertExpression(&testCases[i].exp)
		if testCases[i].exp.Status == "in queue" {
			ExpressionsChan <- testCases[i].exp
		}
	}

	// waiting for calculation
	for len(ExpressionsChan) != 0 || workersCalculating() {

	}
	for _, testCase := range testCases {
		exp := database.GetExpressionById(testCase.exp.Id, testCase.exp.Login)
		if testCase.expectedStatus != exp.Status {
			t.Fatalf("incorrect status: expected: %v, but got: %v", testCase.expectedStatus, exp.Status)
		}
		if testCase.expectedErr != exp.Err {
			t.Fatalf("incorrect error: expected: %v, but got: %v", testCase.expectedErr, exp.Err)
		}
		if testCase.expectedResult != exp.Result {
			t.Fatalf("incorrect result: expected: %v, but got: %v", testCase.expectedResult, exp.Result)
		}
	}
}
