package stats

import (
	"fmt"
	"net"
	"sync"
)

var (
	mutex sync.RWMutex
	stats []string
)

func Counter(key string, val int) {
	AddStat("%s:%d|c", key, val)
}

func Sample(key string, val int, interval float64) {
	AddStat("%s:%d|c|@%f", key, val, interval)
}

func Timer(key string, val int) {
	AddStat("%s:%d|ms", key, val)
}

func Gauge(key string, val int) {
	AddStat("%s:%d|g", key, val)
}

func Set(key string, val int) {
	AddStat("%s:%d|s", key, val)
}

func AddStat(format string, params ...interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	stats = append(stats, Sprintf(format, params...))
}

func Flush() error {
	mutex.RLock()
	defer mutex.Unlock()

	// do stuff

	stats = []string{}
	return nil
}
