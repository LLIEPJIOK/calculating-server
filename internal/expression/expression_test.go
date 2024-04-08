package expression

import (
	"math"
	"testing"
	"time"
)

type expressionTestCase struct {
	name             string
	exp              Expression
	expectedStatus   string
	expectedRpn      []interface{}
	expectedResult   float64
	expectedDuration time.Duration
}

func interfacesSliceCompare(first, second []interface{}) bool {
	if len(first) != len(second) {
		return false
	}
	for i := 0; i < len(first); i++ {
		switch firstVal := first[i].(type) {
		case float64:
			secondVal, ok := second[i].(float64)
			if !ok {
				return false
			}
			if firstVal != secondVal {
				return false
			}
		case uint8:
			secondVal, ok := second[i].(uint8)
			if !ok {
				return false
			}
			if firstVal != secondVal {
				return false
			}
		}
	}
	return true
}

func testExpressionTestCase(t *testing.T, testCase expressionTestCase) {
	t.Parallel()
	if testCase.expectedStatus != testCase.exp.Status {
		t.Fatalf("status: expected: %v, but got: %v", testCase.expectedStatus, testCase.exp.Status)
	}
	if !interfacesSliceCompare(testCase.expectedRpn, testCase.exp.rpn) {
		t.Fatalf("rpn: expected: %v, but got: %v", testCase.expectedRpn, testCase.exp.rpn)
	}

	startTime := time.Now()
	testCase.exp.Calculate()
	endTime := time.Now()
	if math.Abs(testCase.expectedResult-testCase.exp.Result) > 1e-5 {
		t.Fatalf("result: expected: %v, but got: %v", testCase.expectedResult, testCase.exp.Result)
	}

	calculateDuration := endTime.Sub(startTime)
	if calculateDuration > testCase.expectedDuration {
		t.Fatalf("time: expected: %v, but got: %v", testCase.expectedDuration, calculateDuration)
	}
}

func TestExpressionSimple(t *testing.T) {
	testCases := []expressionTestCase{
		{
			name:             "simple addition",
			exp:              New("", "2 + 2"),
			expectedStatus:   "",
			expectedRpn:      []interface{}{2, 2, '+'},
			expectedResult:   4,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple subtraction",
			exp:              New("", "10 - 123"),
			expectedStatus:   "",
			expectedRpn:      []interface{}{10, 123, '-'},
			expectedResult:   -113,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple multiplication",
			exp:              New("", "1.5 * 1.6"),
			expectedStatus:   "",
			expectedRpn:      []interface{}{1.5, 1.6, '*'},
			expectedResult:   2.4,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple division",
			exp:              New("", "1 / 4"),
			expectedStatus:   "",
			expectedRpn:      []interface{}{1, 4, '/'},
			expectedResult:   0.25,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple sign changing",
			exp:              New("", "-5"),
			expectedStatus:   "",
			expectedRpn:      []interface{}{5, '_'},
			expectedResult:   -5,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testExpressionTestCase(t, testCase)
		})
	}
}

func TestExpressionErrors(t *testing.T) {
	testCases := []expressionTestCase{
		{
			name:           "two number in a row",
			exp:            New("", "2 2 + 2"),
			expectedStatus: "invalid expression: two numbers in a row",
		},
		{
			name:           "invalid number",
			exp:            New("", "2;24 * 2"),
			expectedStatus: "invalid expression: strconv.ParseFloat: parsing \"2;24\": invalid syntax",
		},
		{
			name:           "parenthesis without expression inside",
			exp:            New("", "24 + () - 1"),
			expectedStatus: "invalid expression: incorrect placement of brackets",
		},
		{
			name:           "only closing parenthesis",
			exp:            New("", ") - 24 * 5"),
			expectedStatus: "invalid expression: incorrect placement of brackets",
		},
		{
			name:           "closing parenthesis after operation",
			exp:            New("", "24 + )"),
			expectedStatus: "invalid expression: incorrect placement of brackets",
		},
		{
			name:           "unknown symbol",
			exp:            New("", "55 + a - 4"),
			expectedStatus: "invalid expression: unknown math symbol: a",
		},
		{
			name:           "multiply after plus",
			exp:            New("", "1 + *"),
			expectedStatus: "invalid expression: incorrect placement of operations",
		},
		{
			name:           "minus after plus",
			exp:            New("", "2 + - 143"),
			expectedStatus: "invalid expression: incorrect placement of operations",
		},
		{
			name:           "division after parenthesis",
			exp:            New("", "5 - 5 + (/5)"),
			expectedStatus: "invalid expression: incorrect placement of operations",
		},
		{
			name:           "empty expression",
			exp:            New("", "    "),
			expectedStatus: "invalid expression: empty expression",
		},
		{
			name:           "extra closing parenthesis",
			exp:            New("", "2 * (1 + 3))"),
			expectedStatus: "invalid expression: incorrect placement of brackets",
		},
		{
			name:           "missed closing parenthesis",
			exp:            New("", "2 * (1 + 3"),
			expectedStatus: "invalid expression: incorrect placement of brackets",
		},
		{
			name:           "missed operation between parenthesis",
			exp:            New("", "2 * (1 + 3)(2 - 8)"),
			expectedStatus: "invalid expression: incorrect placement of brackets",
		},
		{
			name:           "operation in the end",
			exp:            New("", "2 * 1 /"),
			expectedStatus: "invalid expression: incorrect placement of operations",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if testCase.expectedStatus != testCase.exp.Status {
				t.Fatalf("status: expected: %v, but got: %v", testCase.expectedStatus, testCase.exp.Status)
			}
		})
	}
}

func TestDivisionByZero(t *testing.T) {
	t.Parallel()
	testCase := expressionTestCase{
		name:           "division by zero",
		exp:            New("", "10 / 0"),
		expectedStatus: "invalid expression: division by zero",
	}

	testCase.exp.Calculate()
	if testCase.expectedStatus != testCase.exp.Status {
		t.Fatalf("status: expected: %v, but got: %v", testCase.expectedStatus, testCase.exp.Status)
	}
}
