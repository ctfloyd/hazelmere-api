package logger

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

const FieldRequestId string = "requestId"

type ZeroLogAdapter struct {
	delegate zerolog.Logger
}

func (l *ZeroLogAdapter) Trace(ctx context.Context, message string) {
	e := addRequestId(ctx, l.delegate.Trace())
	e.Msg(message)
}

func (l *ZeroLogAdapter) Debug(ctx context.Context, message string) {
	e := addRequestId(ctx, l.delegate.Debug())
	e.Msg(message)
}

func (l *ZeroLogAdapter) Info(ctx context.Context, message string) {
	e := addRequestId(ctx, l.delegate.Info())
	e.Msg(message)
}

func (l *ZeroLogAdapter) InfoArgs(ctx context.Context, message string, args ...any) {
	l.Info(ctx, fmt.Sprintf(message, args...))
}

func (l *ZeroLogAdapter) Warn(ctx context.Context, message string) {
	e := addRequestId(ctx, l.delegate.Warn())
	e.Msg(message)
}

func (l *ZeroLogAdapter) Error(ctx context.Context, message string) {
	e := addRequestId(ctx, l.delegate.Error())
	e.Msg(message)
}

func NewZeroLogAdapater(level LogLevel) *ZeroLogAdapter {
	delegate := zerolog.New(
		zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = time.RFC3339
			w.PartsOrder = []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				FieldRequestId,
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			}
			w.FieldsExclude = []string{
				FieldRequestId,
			}
		}),
	).Level(convertLogLevel(level)).With().Timestamp().CallerWithSkipFrameCount(4).Logger()
	delegate.Hook()
	return &ZeroLogAdapter{delegate}
}

func convertLogLevel(level LogLevel) zerolog.Level {
	if level == LogLevelTrace {
		return zerolog.TraceLevel
	} else if level == LogLevelDebug {
		return zerolog.DebugLevel
	} else if level == LogLevelInfo {
		return zerolog.InfoLevel
	} else if level == LogLevelWarn {
		return zerolog.WarnLevel
	} else if level == LogLevelError {
		return zerolog.ErrorLevel
	} else {
		return zerolog.InfoLevel
	}
}

func addRequestId(ctx context.Context, e *zerolog.Event) *zerolog.Event {
	requestIdAny := ctx.Value(middleware.RequestIDKey)
	requestId := "NO-REQUEST-ID"
	if requestIdAny != nil && reflect.TypeOf(requestIdAny).Kind() == reflect.String {
		requestId = requestIdAny.(string)
	}

	e.Str(FieldRequestId, requestId)
	return e
}
