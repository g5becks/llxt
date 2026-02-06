package errs

const (
	// CodeNotFound indicates a resource was not found.
	CodeNotFound = "not_found"
	// CodeInvalidEntry indicates an invalid registry entry.
	CodeInvalidEntry = "invalid_entry"

	// CodeNetworkFailure indicates a network operation failed.
	CodeNetworkFailure = "network_failure"
	// CodeRateLimited indicates the request was rate limited.
	CodeRateLimited = "rate_limited"
	// CodeTimeout indicates the request timed out.
	CodeTimeout = "timeout"

	// CodeConfigLoad indicates configuration loading failed.
	CodeConfigLoad = "config_load"
	// CodeConfigParse indicates configuration parsing failed.
	CodeConfigParse = "config_parse"
	// CodeConfigWrite indicates configuration writing failed.
	CodeConfigWrite = "config_write"

	// CodeRepoNotFound indicates a GitHub repository was not found.
	CodeRepoNotFound = "repo_not_found"
	// CodePathNotFound indicates a path in a repository was not found.
	CodePathNotFound = "path_not_found"
	// CodeAPIError indicates a GitHub API error occurred.
	CodeAPIError = "api_error"
)
