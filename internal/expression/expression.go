package expression

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Expression struct {
	Login           string
	Id              uint64
	Exp             string
	Result          float64
	Status          string
	CreationTime    time.Time
	CalculationTime time.Time
	OperationsTimes map[string]uint64
	rpn             []interface{}
}

func New(login, exp string) Expression {
	expression := Expression{
		Login:        login,
		Exp:          exp,
		CreationTime: time.Now(),
		OperationsTimes: map[string]uint64{
			"time-plus":     200,
			"time-minus":    200,
			"time-multiply": 200,
			"time-divide":   200,
		},
	}
	expression.Parse()
	return expression
}

func isDigit(ch uint8) bool {
	return ch >= '0' && ch <= '9'
}

// parsing to reverse polish notation
func (exp *Expression) Parse() {
	st := make([]uint8, 0)
	var prevChar uint8 = '('
	bracketsCnt := 0
	for i := 0; i < len(exp.Exp); i++ {
		ch := exp.Exp[i]
		if isDigit(ch) {
			if isDigit(prevChar) {
				exp.Status = "invalid expression: two numbers in a row"
				return
			}
			length := strings.IndexAny(exp.Exp[i:], "-+*/() \t")
			if length == -1 {
				length = len(exp.Exp) - i
			}
			numb, err := strconv.ParseFloat(exp.Exp[i:i+length], 64)
			if err != nil {
				exp.Status = fmt.Sprintf("invalid expression: %v", err)
				return
			}
			exp.rpn = append(exp.rpn, numb)
			i += length - 1
		} else if ch == '(' {
			if prevChar == ')' {
				exp.Status = "invalid expression: incorrect placement of brackets"
				return
			}
			bracketsCnt++
			st = append(st, ch)
		} else if ch == ')' {
			for len(st) != 0 && st[len(st)-1] != '(' {
				exp.rpn = append(exp.rpn, st[len(st)-1])
				st = st[:len(st)-1]
			}
			if len(st) == 0 || !isDigit(prevChar) {
				exp.Status = "invalid expression: incorrect placement of brackets"
				return
			}
			bracketsCnt--
			st = st[:len(st)-1]
		} else if strings.Contains(" \t", string(ch)) {
			continue
		} else if !strings.Contains("-+*/()", string(ch)) {
			exp.Status = fmt.Sprintf("invalid expression: unknown math symbol: %c", ch)
			return
		} else {
			if !isDigit(prevChar) && !(prevChar == '(' && (ch == '-' || ch == '+')) {
				exp.Status = "invalid expression: incorrect placement of operations"
				return
			}
			if len(st) != 0 {
				for len(st) != 0 {
					top := st[len(st)-1]
					if top == '(' || ((top == '-' || top == '+') && (ch == '*' || ch == '/')) {
						break
					}
					exp.rpn = append(exp.rpn, top)
					st = st[:len(st)-1]
				}
			}
			if prevChar == '(' && (ch == '-' || ch == '+') {
				if ch == '-' {
					st = append(st, '_')
				}
			} else {
				st = append(st, ch)
			}
		}
		prevChar = ch
	}
	for len(st) != 0 {
		exp.rpn = append(exp.rpn, st[len(st)-1])
		st = st[:len(st)-1]
	}
	if len(exp.rpn) == 0 {
		exp.Status = "invalid expression: empty expression"
		return
	} else if bracketsCnt != 0 {
		exp.Status = "invalid expression: incorrect placement of brackets"
		return
	} else if !isDigit(prevChar) && prevChar != ')' {
		exp.Status = "invalid expression: incorrect placement of operations"
		return
	}
}

// calculation from reverse polish notation
func (exp *Expression) Calculate() {
	st := make([]float64, 0)
	for _, v := range exp.rpn {
		if numb, ok := v.(float64); ok {
			st = append(st, numb)
		} else {
			switch v.(uint8) {
			case '_':
				st[len(st)-1] = multiply(st[len(st)-1], -1, exp.OperationsTimes["time-multiply"])
			case '-':
				st[len(st)-2] = minus(st[len(st)-2], st[len(st)-1], exp.OperationsTimes["time-minus"])
				st = st[:len(st)-1]
			case '+':
				st[len(st)-2] = add(st[len(st)-2], st[len(st)-1], exp.OperationsTimes["time-plus"])
				st = st[:len(st)-1]
			case '*':
				st[len(st)-2] = multiply(st[len(st)-2], st[len(st)-1], exp.OperationsTimes["time-multiply"])
				st = st[:len(st)-1]
			case '/':
				if st[len(st)-1] == 0 {
					exp.Status = "invalid expression: division by zero"
					return
				}
				st[len(st)-2] = divide(st[len(st)-2], st[len(st)-1], exp.OperationsTimes["time-divide"])
				st = st[:len(st)-1]
			}
		}
	}
	exp.CalculationTime = time.Now()
	exp.Status = "calculated"
	exp.Result = st[0]
}

func (exp Expression) String() string {
	str := fmt.Sprintf("Id: `%d`, Expression: `%s`, Creation time: `%s`, Status: `%s`",
		exp.Id, exp.Exp, exp.CreationTime.Format("2006-01-02 15:04:05"), exp.Status)
	if exp.Status == "calculated" {
		str += fmt.Sprintf(",  Result: `%v`, Calculation time: `%s`", exp.Result, exp.CalculationTime.Format("2006-01-02 15:04:05"))
	}
	return str
}

func add(a, b float64, seconds uint64) float64 {
	time.Sleep(time.Duration(seconds) * time.Millisecond)
	return a + b
}

func minus(a, b float64, seconds uint64) float64 {
	time.Sleep(time.Duration(seconds) * time.Millisecond)
	return a - b
}

func multiply(a, b float64, seconds uint64) float64 {
	time.Sleep(time.Duration(seconds) * time.Millisecond)
	return a * b
}

func divide(a, b float64, seconds uint64) float64 {
	time.Sleep(time.Duration(seconds) * time.Millisecond)
	return a / b
}

func GetOperationTime(operationsTime map[string]uint64, operation string) uint64 {
	return operationsTime[operation]
}
