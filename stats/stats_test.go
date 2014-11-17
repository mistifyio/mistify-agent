package stats_test

import (
	"github.com/mistifyio/mistify-agent/stats"
	"testing"
)

func TestStats(t *testing.T) {
	go stats.Send("localhost:8125")

	stats.Counter("test", 1)
	stats.Sample("test", 1, 100)
	stats.Gauge("test", 1)
	stats.Set("test", 1)

	timer := stats.StartTimer("test")
	timer.Stop()

	stats.CloseChannel()
}
