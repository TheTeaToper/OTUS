package logger

import (
	"fmt"
	"strings"
	"time"
)

type Logger struct {
	level int
}

const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

func New(level string) *Logger {
	lvl := InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		lvl = DebugLevel
	case "info":
		lvl = InfoLevel
	case "warning":
		lvl = WarningLevel
	case "error":
		lvl = ErrorLevel
	}
	return &Logger{level: lvl}
}

func (l *Logger) log(level string, msg string) {
	fmt.Printf("[%s] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), level, msg)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= DebugLevel {
		l.log("DEBUG", fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= InfoLevel {
		l.log("INFO", fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) Warning(msg string, args ...interface{}) {
	if l.level <= WarningLevel {
		l.log("WARNING", fmt.Sprintf(msg, args...))
	}
}

func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= ErrorLevel {
		l.log("ERROR", fmt.Sprintf(msg, args...))
	}
}
