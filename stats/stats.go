package stats

import (
	"fmt"
	"github.com/mistifyio/mistify-agent/log"
	"net"
)

var stats = make(chan string)

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
	stats <- fmt.Sprintf(format, params...)
}

func Send(addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for stat := range stats {
		_, err := conn.Write([]byte(stat))
		if err != nil {
			log.Error("Couldn't send to statsd: %s. Reconnecting.\n", err.Error())

			conn, err = net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				log.Error("Couldn't reconnect to statsd: %s.\n", err.Error())
			}
		}
	}

	return nil
}
