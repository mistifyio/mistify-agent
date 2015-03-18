package rpc

import (
	"io"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
)

type (
	logger struct {
		handler http.Handler
	}

	logHandler struct {
		http.ResponseWriter
		status int
		size   int
	}
)

func newLogger(w io.Writer, h http.Handler) *logger {
	return &logger{
		handler: h,
	}
}

func (l *logHandler) Write(b []byte) (int, error) {
	if l.status == 0 {
		// The status will be StatusOK if WriteHeader has not been called yet
		l.status = http.StatusOK
	}
	size, err := l.ResponseWriter.Write(b)
	l.size += size
	return size, err
}

func (l *logHandler) WriteHeader(s int) {
	l.ResponseWriter.WriteHeader(s)
	l.status = s
}

func (l *logHandler) Header() http.Header {
	return l.ResponseWriter.Header()
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := logHandler{
		ResponseWriter: w,
	}
	now := time.Now()
	l.handler.ServeHTTP(&h, r)

	method := ""
	m := context.Get(r, RPCPath)
	if m != nil {
		method = m.(string)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		host = r.RemoteAddr
	}

	log.WithFields(log.Fields{
		"method":    method,
		"duration":  time.Since(now).Seconds(),
		"url":       r.URL.RequestURI(),
		"status":    h.status,
		"size":      h.size,
		"userAgent": r.UserAgent(),
		"client":    host,
	}).Info("request processed")
}
