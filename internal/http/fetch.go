package httpclient

import (
	"resty.dev/v3"

	errs "github.com/g5becks/llxt/internal/errors"
)

const (
	statusNotFound    = 404
	statusRateLimited = 429
	statusBadRequest  = 400
)

// Fetcher handles fetching llms.txt content.
type Fetcher struct {
	client *resty.Client
}

// NewFetcher creates a new Fetcher.
func NewFetcher(cfg *Config) *Fetcher {
	return &Fetcher{
		client: NewClient(cfg),
	}
}

// Close closes the underlying HTTP client.
func (f *Fetcher) Close() error {
	return f.client.Close()
}

// Fetch retrieves content from a URL.
func (f *Fetcher) Fetch(url string) (string, error) {
	resp, err := f.client.R().Get(url)
	if err != nil {
		return "", errs.HTTPErr.
			Code(errs.CodeNetworkFailure).
			With("url", url).
			Wrapf(err, "failed to fetch content")
	}

	if resp.StatusCode() == statusNotFound {
		return "", errs.HTTPErr.
			Code(errs.CodeNotFound).
			With("url", url).
			With("status", resp.StatusCode()).
			Errorf("resource not found")
	}

	if resp.StatusCode() == statusRateLimited {
		return "", errs.HTTPErr.
			Code(errs.CodeRateLimited).
			With("url", url).
			With("retry_after", resp.Header().Get("Retry-After")).
			Hint("Wait before retrying or use a different source").
			Errorf("rate limited by server")
	}

	if resp.StatusCode() >= statusBadRequest {
		return "", errs.HTTPErr.
			Code(errs.CodeNetworkFailure).
			With("url", url).
			With("status", resp.StatusCode()).
			Errorf("HTTP error: %d", resp.StatusCode())
	}

	return resp.String(), nil
}

// FetchLLMsTxt fetches llms.txt or llms-full.txt based on full flag.
func (f *Fetcher) FetchLLMsTxt(url string, fullURL *string, full bool) (string, error) {
	targetURL := url
	if full && fullURL != nil {
		targetURL = *fullURL
	}
	return f.Fetch(targetURL)
}
