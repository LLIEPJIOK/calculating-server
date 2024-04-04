package expression

import "time"

var (
	operationTimes = map[string]int64{
		"time-plus":     200,
		"time-minus":    200,
		"time-multiply": 200,
		"time-divide":   200,
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

func UpdateOperationTimes(timePlus, timeMinus, timeMultiply, timeDivide int64) {
	operationTimes["time-plus"] = timePlus
	operationTimes["time-minus"] = timeMinus
	operationTimes["time-multiply"] = timeMultiply
	operationTimes["time-divide"] = timeDivide
}

func add(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time-plus"]) * time.Millisecond)
	return a + b
}

func minus(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time-minus"]) * time.Millisecond)
	return a - b
}

func multiply(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time-multiply"]) * time.Millisecond)
	return a * b
}

func divide(a, b float64) float64 {
	time.Sleep(time.Duration(operationTimes["time-divide"]) * time.Millisecond)
	return a / b
}

func UpdateOperationTimesMap(newOperationTimes map[string]int64) {
	for operation, time := range newOperationTimes {
		operationTimes[operation] = time
	}
}
