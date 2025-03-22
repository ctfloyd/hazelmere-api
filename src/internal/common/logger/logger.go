package logger

import (
	"context"
	"strings"
)

type Logger interface {
	Trace(context.Context, string)
	Debug(context.Context, string)
	Info(context.Context, string)
	InfoArgs(context.Context, string, ...any)
	Warn(context.Context, string)
	Error(context.Context, string)
}

type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

func LogLevelFromString(level string) LogLevel {
	level = strings.ToLower(level)
	if level == "trace" {
		return LogLevelTrace
	} else if level == "debug" {
		return LogLevelDebug
	} else if level == "info" {
		return LogLevelInfo
	} else if level == "warn" {
		return LogLevelWarn
	} else if level == "service_error" {
		return LogLevelError
	} else {
		return LogLevelInfo
	}
}
