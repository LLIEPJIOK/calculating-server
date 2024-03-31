package threadInfo

import (
	"fmt"
	"sync"
	"time"
)

type ThreadInfo struct {
	LastPing time.Time
	Status   string
	Id       int
	Mutex    sync.RWMutex
}

func (ti *ThreadInfo) UpdateStatus(newStatus string) {
	ti.Mutex.Lock()
	defer ti.Mutex.Unlock()
	ti.Status = newStatus
	ti.LastPing = time.Now()
}

func (ti *ThreadInfo) String() string {
	ti.Mutex.RLock()
	defer ti.Mutex.RUnlock()
	return fmt.Sprintf(`Agent #%d: %s. Last ping: %v`, ti.Id, ti.Status, ti.LastPing.Format("2006-01-02 "))
}
