// Package httpclient provides HTTP client functionality with retry and circuit breaker.
package httpclient

import "time"

// Config holds HTTP client configuration.
type Config struct {
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
	Verbose          bool
}

const (
	defaultTimeout          = 30 * time.Second
	defaultRetryCount       = 3
	defaultRetryWaitTime    = 100 * time.Millisecond
	defaultRetryMaxWaitTime = 2 * time.Second
)

// DefaultConfig returns default HTTP configuration.
func DefaultConfig() *Config {
	return &Config{
		Timeout:          defaultTimeout,
		RetryCount:       defaultRetryCount,
		RetryWaitTime:    defaultRetryWaitTime,
		RetryMaxWaitTime: defaultRetryMaxWaitTime,
		Verbose:          false,
	}
}
