package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type contextKey string

const (
	depthKey     contextKey = "call_depth"
	loggerKey    contextKey = "logger"
	maxCallDepth            = 64
)

// WithCallDepth returns a new context with the call depth incremented.
// It panics if the call depth exceeds maxCallDepth.
func WithCallDepth(ctx context.Context) context.Context {
	depth := 0
	if d, ok := ctx.Value(depthKey).(int); ok {
		depth = d
	}

	if depth >= maxCallDepth {
		panic(fmt.Sprintf("call depth exceeded maximum of %d", maxCallDepth))
	}

	return context.WithValue(ctx, depthKey, depth+1)
}

// GetCallDepth returns the current call depth from the context.
func GetCallDepth(ctx context.Context) int {
	if depth, ok := ctx.Value(depthKey).(int); ok {
		return depth
	}
	return 0
}

// ContextWithLogger returns a new context with the provided logger.
func ContextWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext returns the logger from the context, or a new default logger if not found.
func LoggerFromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return New() // Fallback to a new logger
}
