package workers

import (
	"fmt"
	"sync"
	"time"

	"github.com/LLIEPJIOK/calculating-server/internal/database"
	"github.com/LLIEPJIOK/calculating-server/internal/expression"
)

var (
	ExpressionsChan = make(chan expression.Expression, 1000)
	numberOfWorkers = 10
	Workers         = make([]*Worker, numberOfWorkers)
)

type Worker struct {
	LastPing time.Time
	Status   string
	Id       int
	Mutex    sync.RWMutex
}

func (threadInfo *Worker) UpdateStatus(newStatus string) {
	threadInfo.Mutex.Lock()
	defer threadInfo.Mutex.Unlock()
	threadInfo.Status = newStatus
	threadInfo.LastPing = time.Now()
}

func (threadInfo *Worker) String() string {
	threadInfo.Mutex.RLock()
	defer threadInfo.Mutex.RUnlock()
	return fmt.Sprintf(`Agent #%d: %s. Last ping: %v`, threadInfo.Id, threadInfo.Status, threadInfo.LastPing.Format("02.01.2006 15:04:05"))
}

func (threadInfo *Worker) Run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			threadInfo.UpdateStatus("Waiting for expression")
		case exp, ok := <-ExpressionsChan:
			if !ok {
				threadInfo.UpdateStatus("Closed")
				return
			}

			threadInfo.UpdateStatus(fmt.Sprintf("Calculation expression #%v", exp.Id))
			exp.Status = "calculating"
			database.UpdateExpressionStatus(&exp)

			exp.Calculate()

			database.UpdateExpressionStatus(&exp)
			database.UpdateExpressionResult(&exp)
			threadInfo.UpdateStatus("Waiting for expression")
		}
	}
}

func CloseExpressionsChan() {
	close(ExpressionsChan)
}

func Initialize() {
	for i := range numberOfWorkers {
		Workers[i] = &Worker{
			LastPing: time.Now(),
			Status:   "Waiting for expression",
			Id:       i + 1,
		}
		go Workers[i].Run()
	}

	expressions := database.GetUncalculatingExpressions()
	for _, expression := range expressions {
		_ = expression.Parse()
		ExpressionsChan <- *expression
	}
}
