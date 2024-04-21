package workers

import (
	"sync"
	"time"

	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/expression"
)

var (
	ExpressionsChan = make(chan expression.Expression, 1000)
	numberOfWorkers = 12
	Workers         = NewSafeWorkersInfo()
)

type WorkerInfo struct {
	LastPing time.Time
	Status   string
}

func (workerInfo *WorkerInfo) UpdateStatus(newStatus string) {
	workerInfo.Status = newStatus
	workerInfo.LastPing = time.Now()
}

func (workerInfo *WorkerInfo) GetFormatTime() string {
	return time.Since(workerInfo.LastPing).Round(time.Second).String()
}

type SafeWorkersInfo struct {
	WorkersSlice []WorkerInfo
	Mutex        *sync.Mutex
}

func NewSafeWorkersInfo() SafeWorkersInfo {
	workers := make([]WorkerInfo, numberOfWorkers)
	for i := 0; i < numberOfWorkers; i++ {
		workers[i] = WorkerInfo{
			LastPing: time.Now(),
			Status:   "Waiting for expression...",
		}
	}
	return SafeWorkersInfo{
		WorkersSlice: workers,
		Mutex:        &sync.Mutex{},
	}
}

func (safeWorkers *SafeWorkersInfo) UpdateStatus(id int, newStatus string) {
	safeWorkers.Mutex.Lock()
	defer safeWorkers.Mutex.Unlock()
	safeWorkers.WorkersSlice[id].UpdateStatus(newStatus)
}

func (safeWorkers *SafeWorkersInfo) GetStatus(id int) string {
	safeWorkers.Mutex.Lock()
	defer safeWorkers.Mutex.Unlock()
	return safeWorkers.WorkersSlice[id].Status
}

func (safeWorkers *SafeWorkersInfo) GetFormatTime(id int) string {
	safeWorkers.Mutex.Lock()
	defer safeWorkers.Mutex.Unlock()
	return safeWorkers.WorkersSlice[id].GetFormatTime()
}

func Run(workerId int) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			Workers.UpdateStatus(workerId, "Waiting for expression...")
		case exp, ok := <-ExpressionsChan:
			if !ok {
				Workers.UpdateStatus(workerId, "Closed")
				return
			}

			Workers.UpdateStatus(workerId, "Calculation expression...")
			exp.Status = "calculating"
			database.UpdateExpressionStatus(&exp)

			exp.Calculate()

			database.UpdateExpressionStatus(&exp)
			database.UpdateExpressionResult(&exp)
			Workers.UpdateStatus(workerId, "Waiting for expression")
		}
	}
}

func CloseExpressionsChan() {
	close(ExpressionsChan)
}

func Initialize() {
	for i := range numberOfWorkers {
		go Run(i)
	}

	expressions := database.GetUncalculatingExpressions()
	for _, expression := range expressions {
		expression.Parse()
		ExpressionsChan <- *expression
	}
}
