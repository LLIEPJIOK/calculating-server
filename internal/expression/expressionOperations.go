package expression

import "time"

var (
	operationTimes = map[string]int64{
		"time_plus":     200,
		"time_minus":    200,
		"time_multiply": 200,
		"time_divide":   200,
	}
)

func SetOperationTime(key string, val int64) {
	operationTimes[key] = val
}

func GetOperationTimes() map[string]int64 {
	return operationTimes
}

func GetOperationTime(key string) int64 {
	return operationTimes[key]
}

func UpdateOperationsTime(timePlus, timeMinus, timeMultiply, timeDivide int64) {
	operationTimes["time_plus"] = timePlus
	operationTimes["time_minus"] = timeMinus
	operationTimes["time_multiply"] = timeMultiply
	operationTimes["time_divide"] = timeDivide
}

func add(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time_plus"]) * time.Millisecond)
	return a + b
}

func minus(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time_minus"]) * time.Millisecond)
	return a - b
}

func multiply(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time_multiply"]) * time.Millisecond)
	return a * b
}

func divide(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time_divide"]) * time.Millisecond)
	return a / b
}
