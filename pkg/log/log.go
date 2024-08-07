package log

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

const (
	RuntimeError = "RuntimeError"
	loggerName   = "tiple-j-bot"
)

type Log interface {
	RuntimeError(ctx context.Context, msg string, err error)
	FatalRuntimeError(ctx context.Context, msg string, err error)
	InfoContext(ctx context.Context, msg string, args ...any)
}

type Logger struct {
	s *slog.Logger
}

func NewLogger() Log {
	return Logger{otelslog.NewLogger(loggerName)}
}

func (l Logger) RuntimeError(ctx context.Context, msg string, err error) {
	l.s.ErrorContext(ctx, msg, RuntimeError, err)
}

func (l Logger) FatalRuntimeError(ctx context.Context, msg string, err error) {
	l.s.ErrorContext(ctx, msg, RuntimeError, err)
}

func (l Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.s.InfoContext(ctx, msg, args...)
}
