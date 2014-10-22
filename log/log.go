package log

import (
	"fmt"
	"log"
	"os"
)

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	CRITICAL
	FATAL
)

var LogLevel = map[string]int{
	"debug":    DEBUG,
	"info":     INFO,
	"warning":  WARNING,
	"error":    ERROR,
	"critical": CRITICAL,
	"fatal":    FATAL,
}

var (
	CurrentLogLevel = INFO
	logger          = log.New(os.Stdout, "", log.LstdFlags)
)

func Printf(level int, message interface{}, params ...interface{}) {
	if level < CurrentLogLevel {
		return
	}

	switch message.(type) {
	case string:
		logger.Printf(message.(string), params...)
	default:
		logger.Printf("%v", message)
		for _, elem := range params {
			logger.Printf("%v", elem)
		}
	}
}

func SetLogLevel(level string) error {
	level_val, ok := LogLevel[level]

	if ok {
		CurrentLogLevel = level_val
	} else {
		return fmt.Errorf("Invalid log level '%s'", level)
	}

	return nil
}

func Debug(message interface{}, params ...interface{}) {
	Printf(DEBUG, message, params...)
}

func Info(message interface{}, params ...interface{}) {
	Printf(INFO, message, params...)
}

func Warning(message interface{}, params ...interface{}) {
	Printf(WARNING, message, params...)
}

func Error(message interface{}, params ...interface{}) {
	Printf(ERROR, message, params...)
}

func Critical(message interface{}, params ...interface{}) {
	Printf(CRITICAL, message, params...)
}

func Fatal(message interface{}, params ...interface{}) {
	Printf(FATAL, message, params...)
	os.Exit(1)
}
