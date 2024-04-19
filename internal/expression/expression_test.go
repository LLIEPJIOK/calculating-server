package expression

import (
	"fmt"
	"math"
	"testing"
	"time"
)

type expressionTestCase struct {
	name             string
	exp              Expression
	expectedStatus   string
	expectedErr      string
	expectedRpn      []interface{}
	expectedResult   float64
	expectedDuration time.Duration
}

func interfacesSliceToString(slice []interface{}) string {
	stringSlice := "["
	for i := 0; i < len(slice); i++ {
		switch val := slice[i].(type) {
		case float64:
			stringSlice += fmt.Sprintf("%v ", val)
		case uint8:
			stringSlice += string(val) + " "
		}
	}
	if len(stringSlice) != 1 {
		stringSlice = stringSlice[:len(stringSlice)-1]
	}
	stringSlice += "]"
	return stringSlice
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
	if !interfacesSliceCompare(testCase.expectedRpn, testCase.exp.rpn) {
		t.Fatalf("rpn: expected: %v, but got: %v", interfacesSliceToString(testCase.expectedRpn), interfacesSliceToString(testCase.exp.rpn))
	}

	startTime := time.Now()
	testCase.exp.Calculate()
	endTime := time.Now()
	if testCase.expectedStatus != testCase.exp.Status {
		t.Fatalf("status: expected: %v, but got: %v", testCase.expectedStatus, testCase.exp.Status)
	}
	if testCase.expectedErr != testCase.exp.Err {
		t.Fatalf("error: expected: %v, but got: %v", testCase.expectedErr, testCase.exp.Err)
	}
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
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{2.0, 2.0, uint8('+')},
			expectedResult:   4,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple subtraction",
			exp:              New("", "10 - 123"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{10.0, 123.0, uint8('-')},
			expectedResult:   -113,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple multiplication",
			exp:              New("", "1.5 * 1.6"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{1.5, 1.6, uint8('*')},
			expectedResult:   2.4,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple division",
			exp:              New("", "1 / 4"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{1.0, 4.0, uint8('/')},
			expectedResult:   0.25,
			expectedDuration: 200*time.Millisecond + 20*time.Millisecond,
		},
		{
			name:             "simple sign changing",
			exp:              New("", "-5"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{5.0, uint8('_')},
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

func TestExpressionSimpleErrors(t *testing.T) {
	testCases := []expressionTestCase{
		{
			name:           "two number in a row",
			exp:            New("", "2 2 + 2"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: two numbers in a row",
		},
		{
			name:           "invalid number",
			exp:            New("", "2;24 * 2"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: strconv.ParseFloat: parsing \"2;24\": invalid syntax",
		},
		{
			name:           "parenthesis without expression inside",
			exp:            New("", "24 + () - 1"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
		},
		{
			name:           "only closing parenthesis",
			exp:            New("", ") - 24 * 5"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
		},
		{
			name:           "closing parenthesis after operation",
			exp:            New("", "24 + )"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
		},
		{
			name:           "unknown symbol",
			exp:            New("", "55 + a - 4"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: unknown math symbol: a",
		},
		{
			name:           "multiply after plus",
			exp:            New("", "1 + *"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of operations",
		},
		{
			name:           "minus after plus",
			exp:            New("", "2 + - 143"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of operations",
		},
		{
			name:           "division after parenthesis",
			exp:            New("", "5 - 5 + (/5)"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of operations",
		},
		{
			name:           "empty expression",
			exp:            New("", "    "),
			expectedStatus: "error",
			expectedErr:    "invalid expression: empty expression",
		},
		{
			name:           "extra closing parenthesis",
			exp:            New("", "2 * (1 + 3))"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
		},
		{
			name:           "missed closing parenthesis",
			exp:            New("", "2 * (1 + 3"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
		},
		{
			name:           "missed operation between parenthesis",
			exp:            New("", "2 * (1 + 3)(2 - 8)"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
		},
		{
			name:           "operation in the end",
			exp:            New("", "2 * 1 /"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of operations",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.exp.Calculate()
			if testCase.expectedStatus != testCase.exp.Status {
				t.Fatalf("status: expected: %v, but got: %v", testCase.expectedStatus, testCase.exp.Status)
			}
			if testCase.expectedErr != testCase.exp.Err {
				t.Fatalf("error: expected: %v, but got: %v", testCase.expectedErr, testCase.exp.Err)
			}
		})
	}
}

func TestDivisionByZero(t *testing.T) {
	testCases := []expressionTestCase{
		{
			name:           "simple division by zero",
			exp:            New("", "10 / 0"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: division by zero",
		},
		{
			name:           "complex division by zero",
			exp:            New("", "10 * 1 - 6 * (9 - 4) / (5 * (5 - 6) + 5)"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: division by zero",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.exp.Calculate()
			if testCase.expectedStatus != testCase.exp.Status {
				t.Fatalf("status: expected: %v, but got: %v", testCase.expectedStatus, testCase.exp.Status)
			}
		})
	}
}

func TestExpression(t *testing.T) {
	testCases := []expressionTestCase{
		{
			name:             "first",
			exp:              New("", "6*(1.5/6+2*(7-8))*2/3"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{6.0, 1.5, 6.0, uint8('/'), 2.0, 7.0, 8.0, uint8('-'), uint8('*'), uint8('+'), uint8('*'), 2.0, uint8('*'), 3.0, uint8('/')},
			expectedResult:   -7,
			expectedDuration: 7*200*time.Millisecond + 150*time.Millisecond,
		},
		{
			name:             "second",
			exp:              New("", "(1272 + 5123) * 52 / (52 * 52) + 52 * 100 - 48"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{1272.0, 5123.0, uint8('+'), 52.0, uint8('*'), 52.0, 52.0, uint8('*'), uint8('/'), 52.0, 100.0, uint8('*'), uint8('+'), 48.0, uint8('-')},
			expectedResult:   5274.98076923,
			expectedDuration: 7*200*time.Millisecond + 150*time.Millisecond,
		},
		{
			name:             "third",
			exp:              New("", "2/3/4/9+2*3*(1-2*9-3+6)/3*7"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{2.0, 3.0, uint8('/'), 4.0, uint8('/'), 9.0, uint8('/'), 2.0, 3.0, uint8('*'), 1.0, 2.0, 9.0, uint8('*'), uint8('-'), 3.0, uint8('-'), 6.0, uint8('+'), uint8('*'), 3.0, uint8('/'), 7.0, uint8('*'), uint8('+')},
			expectedResult:   -195.9814814,
			expectedDuration: 12*200*time.Millisecond + 150*time.Millisecond,
		},
		{
			name:             "fourth",
			exp:              New("", "1.5315*0.2165+(-652.18*8)/1.2"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{1.5315, 0.2165, uint8('*'), 652.18, uint8('_'), 8.0, uint8('*'), 1.2, uint8('/'), uint8('+')},
			expectedResult:   -4347.5350969,
			expectedDuration: 5*200*time.Millisecond + 150*time.Millisecond,
		},
		{
			name:             "fifth",
			exp:              New("", "(7863 + 10236) * 123 / (2011 * 9) + 52 * 989 - 551"),
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{7863.0, 10236.0, uint8('+'), 123.0, uint8('*'), 2011.0, 9.0, uint8('*'), uint8('/'), 52.0, 989.0, uint8('*'), uint8('+'), 551.0, uint8('-')},
			expectedResult:   51000,
			expectedDuration: 7*200*time.Millisecond + 150*time.Millisecond,
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
			name:           "incorrect placement of brackets",
			exp:            New("", "6*(1.5/6+2*(7-8)*2/3"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of parenthesis",
		},
		{
			name:           "unknown symbol",
			exp:            New("", "(1272 + 5123) * 52 / (52 x 52) + 52 * 100 - 48"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: unknown math symbol: x",
		},
		{
			name:           "two divisions in a row",
			exp:            New("", "2/3/4//9+2*3"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: incorrect placement of operations",
		},
		{
			name:           "fourth",
			exp:            New("", "1.5315*0.2165+(-652.18*8o)/1.2"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: strconv.ParseFloat: parsing \"8o\": invalid syntax",
		},
		{
			name:           "fifth",
			exp:            New("", "(7863 + 10236 5.5) * 123 / (2011 * 9) + 52 * 989 - 551"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: two numbers in a row",
		},
		{
			name:           "sixth",
			exp:            New("", "(7863 + 1023.5.5) * 123 / (2011 * 9) + 52 * 989 - 551"),
			expectedStatus: "error",
			expectedErr:    "invalid expression: strconv.ParseFloat: parsing \"1023.5.5\": invalid syntax",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.expectedStatus != testCase.exp.Status {
				t.Fatalf("status: expected: %v, but got: %v", testCase.expectedStatus, testCase.exp.Status)
			}
			if testCase.expectedErr != testCase.exp.Err {
				t.Fatalf("error: expected: %v, but got: %v", testCase.expectedErr, testCase.exp.Err)
			}
		})
	}
}

func TestExpressionCalculationTime(t *testing.T) {
	testCases := []expressionTestCase{
		{
			name: "first",
			exp: Expression{
				Exp: "34.5/6*(7-88)",
				OperationsTimes: map[string]uint64{
					"time-plus":     100,
					"time-minus":    500,
					"time-multiply": 200,
					"time-divide":   1000,
				},
			},
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{34.5, 6.0, uint8('/'), 7.0, 88.0, uint8('-'), uint8('*')},
			expectedResult:   -465.75,
			expectedDuration: 1700*time.Millisecond + 150*time.Millisecond,
		},
		{
			name: "second",
			exp: Expression{
				Exp: "52 / (52 * 52) + 52 * (52 - 52)",
				OperationsTimes: map[string]uint64{
					"time-plus":     300,
					"time-minus":    400,
					"time-multiply": 700,
					"time-divide":   100,
				},
			},
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{52.0, 52.0, 52.0, uint8('*'), uint8('/'), 52.0, 52.0, 52.0, uint8('-'), uint8('*'), uint8('+')},
			expectedResult:   0.01923076,
			expectedDuration: 2200*time.Millisecond + 150*time.Millisecond,
		},
		{
			name: "third",
			exp: Expression{
				Exp: "1+2-3*4/5",
				OperationsTimes: map[string]uint64{
					"time-plus":     0,
					"time-minus":    0,
					"time-multiply": 0,
					"time-divide":   0,
				},
			},
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{1.0, 2.0, uint8('+'), 3.0, 4.0, uint8('*'), 5.0, uint8('/'), uint8('-')},
			expectedResult:   0.6,
			expectedDuration: 0*time.Millisecond + 150*time.Millisecond,
		},
		{
			name: "fourth",
			exp: Expression{
				Exp: "(1.8/9.5+4.5-0-0-0*5)",
				OperationsTimes: map[string]uint64{
					"time-plus":     100,
					"time-minus":    200,
					"time-multiply": 300,
					"time-divide":   400,
				},
			},
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{1.8, 9.5, uint8('/'), 4.5, uint8('+'), 0, uint8('-'), 0, uint8('-'), 0, 5.0, uint8('*'), uint8('-')},
			expectedResult:   4.68947368,
			expectedDuration: 1400*time.Millisecond + 150*time.Millisecond,
		},
		{
			name: "fifth",
			exp: Expression{
				Exp: "(2 + 111)  / (2011 - 9) * 55",
				OperationsTimes: map[string]uint64{
					"time-plus":     500,
					"time-minus":    500,
					"time-multiply": 500,
					"time-divide":   500,
				},
			},
			expectedStatus:   "calculated",
			expectedErr:      "",
			expectedRpn:      []interface{}{2.0, 111.0, uint8('+'), 2011.0, 9.0, uint8('-'), uint8('/'), 55.0, uint8('*')},
			expectedResult:   3.1043956,
			expectedDuration: 2500*time.Millisecond + 150*time.Millisecond,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.exp.Parse()
			testExpressionTestCase(t, testCase)
		})
	}
}
