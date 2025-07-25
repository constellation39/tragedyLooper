package logger

import (
	"context"

	"go.uber.org/zap"
)

// loggerKey is an unexported type to be used as the key for storing the
// logger in the context.Context. This prevents collisions with keys from
// other packages.
type loggerKey struct{}

// WithContext returns a new context with the provided logger embedded.
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext retrieves the logger from the context. If no logger is
// found, it returns a new no-op logger, which does nothing.
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return logger
	}
	// Return a no-op logger so that calls to it don't panic.
	return zap.NewNop()
}
