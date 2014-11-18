package stats

import (
	"time"
)

type TimerMetric struct {
	start time.Time
	key   string
}

func StartTimer(key string) *TimerMetric {
	timer := TimerMetric{start: time.Now(), key: key}
	return &timer
}

func (timer *TimerMetric) Stop() {
	interval := time.Since(timer.start)
	Timer(timer.key, int(interval.Seconds()*1000))
}
