package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/constellation39/tragedyLooper/internal/logger"

	"github.com/stretchr/testify/assert"
)

func TestCallDepthExceeded(t *testing.T) {
	// Create a new logger and a dummy server for testing.
	log := logger.New()
	srv := NewServer("", nil, log)

	// Create a handler that will recursively call a function to exceed the call depth.
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var recursiveFunc func(ctx context.Context)
		recursiveFunc = func(ctx context.Context) {
			newCtx := logger.WithCallDepth(ctx)
			recursiveFunc(newCtx)
		}

		// We expect this to panic.
		assert.Panics(t, func() {
			recursiveFunc(ctx)
		})
	})

	// Create a test request and response recorder.
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Wrap the handler with the logging middleware to initialize the call depth.
	middleware := srv.LoggingMiddleware(h)
	middleware.ServeHTTP(rr, req)
}
