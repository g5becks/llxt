package httpclient

import (
	"log/slog"
	"os"
	"time"

	"resty.dev/v3"
)

const (
	cbFailureThreshold = 3
	cbSuccessThreshold = 1
	cbResetTimeout     = 30 * time.Second
)

//nolint:gochecknoglobals // logger is shared across package
var logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

// NewClient creates a new HTTP client with retry and circuit breaker.
func NewClient(cfg *Config) *resty.Client {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create circuit breaker
	cb := resty.NewCircuitBreakerWithCount(
		cbFailureThreshold,
		cbSuccessThreshold,
		cbResetTimeout,
		resty.CircuitBreaker5xxPolicy,
	).OnStateChange(func(oldState, newState resty.CircuitBreakerState) {
		logger.Warn("circuit breaker state changed",
			slog.Any("from", oldState),
			slog.Any("to", newState),
		)
	})

	client := resty.New().
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(cfg.RetryWaitTime).
		SetRetryMaxWaitTime(cfg.RetryMaxWaitTime).
		SetCircuitBreaker(cb).
		AddRetryHooks(func(res *resty.Response, _ error) {
			logger.Debug("retrying request",
				slog.String("url", res.Request.URL),
				slog.Int("attempt", res.Request.Attempt),
			)
		})

	if cfg.Verbose {
		client.SetDebug(true)
	}

	return client
}
