package logger

import (
	"fmt"
	//"log"
)

type LogLevel int

const (
	LogLevelFatal LogLevel = iota // 0
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug // 4
)

var (
	//logger  = log.New(*log.Logger
	logLevel = LogLevelDebug
)

func logPrint(level LogLevel, format string, values ...interface{}) {
	if level <= logLevel {
		//log.Printf(format, values)
		fmt.Printf(format, values...)
		fmt.Println()
	}
}

func Fatal(format string, values ...interface{}) {
	logPrint(LogLevelFatal, format, values...)
}

func Error(format string, values ...interface{}) {
	logPrint(LogLevelError, format, values...)
}

func Warn(format string, values ...interface{}) {
	logPrint(LogLevelWarn, format, values...)
}

func Info(format string, values ...interface{}) {
	logPrint(LogLevelInfo, format, values...)
}

func Debug(format string, values ...interface{}) {
	logPrint(LogLevelDebug, format, values...)
}
