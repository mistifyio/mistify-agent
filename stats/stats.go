package stats

import (
	"fmt"
	"net"
)

var (
	stats      = make(chan string)
	remoteAddr *net.TCPAddr
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
	stats <- fmt.Sprintf(format, params...)
}

func Flush() error {
	conn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for stat := range stats {
		_, err := conn.Write([]byte(stat))
		if err != nil {
			return nil
		}
	}

	return nil
}
