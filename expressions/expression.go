package expressions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	operationsTime = map[string]int64{
		"time_plus":     200,
		"time_minus":    200,
		"time_multiply": 200,
		"time_divide":   200,
	}
)

type Expression struct {
	Id              int64
	Exp             string
	Result          float64
	Status          string
	CreationTime    time.Time
	CalculationTime time.Time
	rpn             []interface{}
}

func NewExpression(exp string) (*Expression, error) {
	expression := &Expression{
		Exp:          exp,
		CreationTime: time.Now(),
	}
	return expression, expression.Parse()
}

func isDigit(ch uint8) bool {
	return ch >= '0' && ch <= '9'
}

// parsing to reverse polish notation
func (exp *Expression) Parse() error {
	st := make([]uint8, 0)
	var prevChar uint8 = '('
	bracketsCnt := 0
	for i := 0; i < len(exp.Exp); i++ {
		ch := exp.Exp[i]
		if isDigit(ch) {
			if isDigit(prevChar) {
				err := errors.New("two numbers in a row")
				exp.Status = fmt.Sprintf("invalid expression: %v", err)
				return err
			}
			length := strings.IndexAny(exp.Exp[i:], "-+*/() \t")
			if length == -1 {
				length = len(exp.Exp) - i
			}
			numb, err := strconv.ParseFloat(exp.Exp[i:i+length], 64)
			if err != nil {
				exp.Status = fmt.Sprintf("invalid expression: %v", err)
				return err
			}
			exp.rpn = append(exp.rpn, numb)
			i += length - 1
		} else if ch == '(' {
			if prevChar == ')' {
				err := errors.New("incorrect placement of brackets")
				exp.Status = fmt.Sprintf("invalid expression: %v", err)
				return err
			}
			bracketsCnt++
			st = append(st, ch)
		} else if ch == ')' {
			for len(st) != 0 && st[len(st)-1] != '(' {
				exp.rpn = append(exp.rpn, st[len(st)-1])
				st = st[:len(st)-1]
			}
			if len(st) == 0 || !isDigit(prevChar) {
				err := errors.New("incorrect placement of brackets")
				exp.Status = fmt.Sprintf("invalid expression: %v", err)
				fmt.Println(prevChar)
				return err
			}
			bracketsCnt--
			st = st[:len(st)-1]
		} else if strings.Contains(" \t", string(ch)) {
			continue
		} else if !strings.Contains("-+*/()", string(ch)) {
			err := fmt.Errorf("unknown math symbol: %v", ch)
			exp.Status = fmt.Sprintf("invalid expression: %v", err)
			return err
		} else {
			if !isDigit(prevChar) && !(prevChar == '(' && (ch == '-' || ch == '+')) {
				err := errors.New("incorrect placement of operations")
				exp.Status = fmt.Sprintf("invalid expression: %v", err)
				return err
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
		err := errors.New("empty expression")
		exp.Status = fmt.Sprintf("invalid expression: %v", err)
		return err
	} else if bracketsCnt != 0 {
		err := errors.New("incorrect placement of brackets")
		exp.Status = fmt.Sprintf("invalid expression: %v", err)
		return err
	} else if !isDigit(prevChar) && prevChar != ')' {
		err := errors.New("incorrect placement of operations")
		exp.Status = fmt.Sprintf("invalid expression: %v", err)
		return err
	}
	return nil
}

func add(a, b float64) float64 {
	time.Sleep(time.Duration(operationsTime["time_plus"]) * time.Millisecond)
	return a + b
}

func minus(a, b float64) float64 {
	time.Sleep(time.Duration(operationsTime["time_minus"]) * time.Millisecond)
	return a - b
}

func multiply(a, b float64) float64 {
	time.Sleep(time.Duration(operationsTime["time_multiply"]) * time.Millisecond)
	return a * b
}

func divide(a, b float64) float64 {
	time.Sleep(time.Duration(operationsTime["time_divide"]) * time.Millisecond)
	return a / b
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
				st[len(st)-1] = multiply(st[len(st)-1], -1)
			case '-':
				st[len(st)-2] = minus(st[len(st)-2], st[len(st)-1])
				st = st[:len(st)-1]
			case '+':
				st[len(st)-2] = add(st[len(st)-2], st[len(st)-1])
				st = st[:len(st)-1]
			case '*':
				st[len(st)-2] = multiply(st[len(st)-2], st[len(st)-1])
				st = st[:len(st)-1]
			case '/':
				st[len(st)-2] = divide(st[len(st)-2], st[len(st)-1])
				st = st[:len(st)-1]
			}
		}
	}
	exp.CalculationTime = time.Now()
	exp.Status = "calculated"
	exp.Result = st[0]
}

func (exp Expression) String() string {
	str := fmt.Sprintf("Id: `%d`, Expression: `%s`, Creation date: `%s`, Status: `%s`",
		exp.Id, exp.Exp, exp.CreationTime.Format("2006-01-02 15:04:05"), exp.Status)
	if exp.Status == "calculated" {
		str += fmt.Sprintf(",  Result: `%v`, Calculation date: `%s`", exp.Result, exp.CalculationTime.Format("2006-01-02 15:04:05"))
	}
	return str
}
