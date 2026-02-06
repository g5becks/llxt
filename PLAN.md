# llxt - LLMs.txt Fetcher & Generator

A Go CLI tool that fetches `llms.txt` files from a bundled registry or generates them on-the-fly from GitHub documentation.

## Overview

```
llxt fetch <name>
    │
    ├─► Check bundled registry
    │   └─► Found? Fetch and return llms.txt
    │
    ├─► Check user config (~/.config/llxt/sources.toml)
    │   └─► Has mapping? Generate from GitHub docs
    │
    └─► Not found? Error with suggestion to add via `llxt add`
```

---

## Core Features

### 1. Bundled Registry (Pre-Scraped)
Ships with a comprehensive registry of **1000+ llms.txt URLs**, already scraped from:
- https://directory.llmstxt.cloud/
- https://llmstxt.site/
- https://llmstxthub.com/

**Existing scraped data files:**
- `directory_entries.json` - Simple format: key, name, domain, llms_url, llms_full_url
- `websites.json` - Rich format: name, domain, description, category, llmsTxtUrl, llmsFullTxtUrl, favicon, publishedAt

These files are embedded into the binary using Go's `embed` package.

### 2. AI-First Output Design
**Primary use case is AI agents**, so output is optimized for machine consumption:
- **Default output**: Raw llms.txt content (AI-friendly)
- **Human output**: Use `--table` or `--json` flags for formatted output
- Progress spinners show during fetch/generation (stderr, not stdout)

### 3. User-Extensible
Users can add their own sources via CLI:
```bash
llxt add vite --url https://vite.dev/llms.txt
llxt add goreleaser --github goreleaser/goreleaser --path www/docs
```

### 3. On-the-Fly Generation
Generate `llms.txt` from GitHub documentation when no registry entry exists.

### 4. No Local Downloads
All operations happen in-memory using GitHub's APIs—no need to clone repos.

---

## Technology Stack

### Dependencies

| Library | Purpose | Version |
|---------|---------|---------|
| [knadh/koanf](https://github.com/knadh/koanf) | Configuration management (TOML) | v2 |
| [urfave/cli](https://github.com/urfave/cli) | CLI framework | v3 |
| [go-resty/resty](https://resty.dev) | HTTP client with retry & circuit breaker | v3 |
| [jedib0t/go-pretty](https://github.com/jedib0t/go-pretty) | CLI tables & formatting | v6 |
| [samber/oops](https://github.com/samber/oops) | Structured error handling | v1 |

### Installation Commands

```bash
# Core dependencies
go get github.com/knadh/koanf/v2
go get github.com/urfave/cli/v3@latest
go get resty.dev/v3
go get github.com/jedib0t/go-pretty/v6
go get github.com/samber/oops

# Koanf providers & parsers
go get github.com/knadh/koanf/providers/file
go get github.com/knadh/koanf/providers/confmap
go get github.com/knadh/koanf/providers/cliflagv3
go get github.com/knadh/koanf/parsers/toml/v2
```

---

## Error Handling Strategy

### Using samber/oops for Structured Errors

All errors should use `oops` for rich context, stack traces, and debugging hints.

#### Error Domains

Define error domains for each component:
```go
const (
    DomainRegistry  = "registry"
    DomainConfig    = "config"
    DomainHTTP      = "http"
    DomainGitHub    = "github"
    DomainGenerator = "generator"
)
```

#### Error Codes

```go
const (
    // Registry errors
    CodeNotFound        = "not_found"
    CodeInvalidEntry    = "invalid_entry"

    // HTTP errors
    CodeNetworkFailure  = "network_failure"
    CodeRateLimited     = "rate_limited"
    CodeTimeout         = "timeout"

    // Config errors
    CodeConfigLoad      = "config_load"
    CodeConfigParse     = "config_parse"
    CodeConfigWrite     = "config_write"

    // GitHub errors
    CodeRepoNotFound    = "repo_not_found"
    CodePathNotFound    = "path_not_found"
    CodeAPIError        = "api_error"
)
```

#### Error Builder Pattern

```go
import "github.com/samber/oops"

// Create domain-specific error builders
var (
    registryErr = oops.In(DomainRegistry).Tags("registry")
    httpErr     = oops.In(DomainHTTP).Tags("http", "network")
    configErr   = oops.In(DomainConfig).Tags("config")
    githubErr   = oops.In(DomainGitHub).Tags("github", "api")
)

// Usage in registry package
func Lookup(name string) (*Entry, error) {
    entry, ok := bundled[name]
    if !ok {
        return nil, registryErr.
            Code(CodeNotFound).
            With("name", name).
            Hint("Use 'llxt list' to see available sources or 'llxt add' to add a new one").
            Errorf("source %q not found in registry", name)
    }
    return entry, nil
}

// Usage in HTTP client
func fetchLLMSTxt(url string) (string, error) {
    resp, err := client.R().Get(url)
    if err != nil {
        return "", httpErr.
            Code(CodeNetworkFailure).
            With("url", url).
            Wrapf(err, "failed to fetch llms.txt")
    }

    if resp.StatusCode() == 429 {
        return "", httpErr.
            Code(CodeRateLimited).
            With("url", url).
            With("retry_after", resp.Header().Get("Retry-After")).
            Hint("Wait before retrying or use a different source").
            Errorf("rate limited by server")
    }

    return resp.String(), nil
}
```

#### Public-Facing Error Messages

```go
// Use .Public() for user-friendly messages
err := httpErr.
    Public("Could not fetch documentation. Please check your internet connection.").
    Code(CodeNetworkFailure).
    Wrapf(innerErr, "TCP connection refused to %s", url)

// Retrieve public message for CLI output
userMessage := oops.GetPublic(err, "An unexpected error occurred")
```

---

## Logging Strategy

### Using log/slog with oops Integration

```go
import (
    "log/slog"
    "github.com/samber/oops"
)

// Global logger setup
var logger *slog.Logger

func initLogger(verbose bool) {
    level := slog.LevelInfo
    if verbose {
        level = slog.LevelDebug
    }

    logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: level,
    }))
}

// Logging errors with oops context
func logError(err error) {
    if oopsErr, ok := oops.AsOops(err); ok {
        logger.Error(err.Error(),
            slog.String("domain", oopsErr.Domain()),
            slog.Any("code", oopsErr.Code()),
            slog.Any("context", oopsErr.Context()),
            slog.String("trace", oopsErr.Trace()),
        )
    } else {
        logger.Error(err.Error())
    }
}
```

### Log Levels

| Level | Usage |
|-------|-------|
| `Debug` | HTTP request/response details, config loading steps |
| `Info` | Command execution, successful operations |
| `Warn` | Recoverable issues, deprecation notices |
| `Error` | Failed operations, non-recoverable errors |

---

## HTTP Client Configuration

### Resty Client Setup with Retry & Circuit Breaker

```go
import (
    "time"
    "resty.dev/v3"
)

func NewHTTPClient(cfg *Config) *resty.Client {
    // Create circuit breaker
    cb := resty.NewCircuitBreakerWithCount(
        3,              // failure threshold
        1,              // success threshold
        30*time.Second, // reset timeout
        resty.CircuitBreaker5xxPolicy,
    ).OnStateChange(func(old, new resty.CircuitBreakerState) {
        logger.Warn("circuit breaker state changed",
            slog.String("from", old.String()),
            slog.String("to", new.String()),
        )
    })

    client := resty.New().
        SetTimeout(cfg.Timeout).
        SetRetryCount(cfg.RetryCount).
        SetRetryWaitTime(cfg.RetryWaitTime).
        SetRetryMaxWaitTime(cfg.RetryMaxWaitTime).
        SetCircuitBreaker(cb).
        AddRetryHooks(func(res *resty.Response, err error) {
            logger.Debug("retrying request",
                slog.String("url", res.Request.URL),
                slog.Int("attempt", res.Request.Attempt),
                slog.String("error", err.Error()),
            )
        })

    // Enable debug mode if verbose
    if cfg.Verbose {
        client.SetDebug(true)
    }

    return client
}
```

### Default HTTP Configuration

```go
type HTTPConfig struct {
    Timeout         time.Duration `koanf:"timeout"`
    RetryCount      int           `koanf:"retry_count"`
    RetryWaitTime   time.Duration `koanf:"retry_wait_time"`
    RetryMaxWaitTime time.Duration `koanf:"retry_max_wait_time"`
}

var DefaultHTTPConfig = HTTPConfig{
    Timeout:          30 * time.Second,
    RetryCount:       3,
    RetryWaitTime:    100 * time.Millisecond,
    RetryMaxWaitTime: 2 * time.Second,
}
```

---

## Registry Design

### Bundled Registry (Using go:embed)

The registry is compiled into the binary using Go's `embed` package. Two data sources are merged:

```go
import "embed"

//go:embed registry/directory_entries.json
var directoryEntriesJSON []byte

//go:embed registry/websites.json
var websitesJSON []byte
```

### Data File Formats

**`directory_entries.json`** - Simple format (~1000+ entries):
```json
[
  {
    "key": "anthropic",
    "name": "anthropic.com",
    "domain": "docs.anthropic.com",
    "llms_url": "https://docs.anthropic.com/llms.txt",
    "llms_full_url": "https://docs.anthropic.com/llms-full.txt"
  }
]
```

**`websites.json`** - Rich format with descriptions & categories:
```json
[
  {
    "name": "Anthropic",
    "domain": "https://anthropic.com",
    "description": "AI safety company building reliable AI systems",
    "llmsTxtUrl": "https://docs.anthropic.com/llms.txt",
    "llmsFullTxtUrl": "https://docs.anthropic.com/llms-full.txt",
    "category": "ai-ml",
    "favicon": "https://www.google.com/s2/favicons?domain=anthropic.com&sz=128",
    "publishedAt": "2025-02-24"
  }
]
```

### Registry Entry Structure (Unified)

```go
type RegistryEntry struct {
    Key         string  `json:"key"`
    Name        string  `json:"name"`
    Domain      string  `json:"domain"`
    Description string  `json:"description,omitempty"`
    Category    string  `json:"category,omitempty"`
    LLMsURL     string  `json:"llms_url"`
    LLMsFullURL *string `json:"llms_full_url,omitempty"` // nil if not available
}
```

### Categories (from websites.json)

The scraped data includes categorization:
- **ai-ml** - AI/ML tools and platforms
- **developer-tools** - SDKs, frameworks, libraries
- **infrastructure-cloud** - Hosting, cloud services
- **integration-automation** - Workflow automation
- **data-analytics** - Data platforms

### User Config

Additional sources stored in `~/.config/llxt/sources.toml`:

```toml
# Direct URL sources (added via llxt add --url)
[registry.effect]
name = "Effect-TS"
url = "https://effect.website/llms.txt"

# GitHub-based sources (for generation)
[sources.goreleaser]
github = "goreleaser/goreleaser"
path = "www/docs"
pattern = "**/*.md"

[sources.drizzle]
github = "drizzle-team/drizzle-orm"
path = "docs"
pattern = "**/*.mdx"
```

---

## CLI Interface

### Output Philosophy: AI-First

Since the primary consumers are AI agents, output is designed for machine consumption by default:

| Command | Default Output | Human-Friendly |
|---------|---------------|----------------|
| `llxt fetch <name>` | Raw llms.txt content | N/A (content is content) |
| `llxt list` | Newline-separated keys | `--table` or `--json` |
| `llxt search <pattern>` | Newline-separated keys | `--table` or `--json` |
| `llxt info <name>` | Key-value pairs | `--table` or `--json` |

**Progress indicators** (spinners, status messages) are written to **stderr**, keeping stdout clean for piping.

### Commands

```bash
# Fetch llms.txt (AI-friendly: outputs raw content to stdout)
llxt fetch <name>
llxt fetch vite                    # Outputs raw llms.txt content
llxt fetch vite --full             # Fetch llms-full.txt if available
llxt fetch vite -o docs/vite.txt   # Save to file

# List all available sources (AI-friendly: one key per line)
llxt list                          # Outputs: vite\nsvelte\ncloudflare\n...
llxt list --table                  # Human-friendly table format
llxt list --json                   # JSON array output
llxt list --category ai-ml         # Filter by category
llxt list --registry               # Only bundled registry
llxt list --user                   # Only user-added sources

# Search sources (AI-friendly: matching keys, one per line)
llxt search <pattern>              # Fuzzy search, outputs matching keys
llxt search react --table          # Human-friendly table with details
llxt search api --json             # JSON array with full entry details

# Get info about a source (AI-friendly: key=value format)
llxt info <name>                   # Outputs: name=Vite\nurl=https://...\n
llxt info vite --json              # Full JSON object
llxt info vite --table             # Human-friendly table

# Add a new source to user config
llxt add <name> --url <url>                          # Direct URL
llxt add <name> --github <owner/repo> --path <path>  # GitHub docs

# Remove a user source
llxt remove <name>

# Registry management (for maintainers)
llxt registry add <name> --url <url> [--full-url <url>] --name "Display Name"
llxt registry remove <name>
llxt registry validate             # Validate all registry URLs (with progress)
llxt registry scrape               # Scrape and update from directories

# Update bundled registry (future: fetch latest from upstream)
llxt update
```

### Exit Codes

```go
const (
    ExitSuccess         = 0
    ExitGeneralError    = 1
    ExitNotFound        = 2
    ExitNetworkError    = 3
    ExitConfigError     = 4
    ExitInvalidInput    = 5
    ExitRateLimited     = 6
)

// Usage with urfave/cli
func fetchAction(ctx context.Context, cmd *cli.Command) error {
    content, err := fetch(name)
    if err != nil {
        code := ExitGeneralError
        if oops.AsError[*NotFoundError](err) != nil {
            code = ExitNotFound
        }
        // Get public message for user
        msg := oops.GetPublic(err, "Failed to fetch llms.txt")
        return cli.Exit(msg, code)
    }
    fmt.Print(content)
    return nil
}
```

### Global Flags

```go
Flags: []cli.Flag{
    &cli.StringFlag{
        Name:    "config",
        Aliases: []string{"c"},
        Usage:   "Path to config file",
        Value:   "~/.config/llxt/sources.toml",
        Sources: cli.EnvVars("LLXT_CONFIG"),
    },
    &cli.StringFlag{
        Name:    "output",
        Aliases: []string{"o"},
        Usage:   "Output file (default: stdout)",
        Sources: cli.EnvVars("LLXT_OUTPUT"),
    },
    &cli.BoolFlag{
        Name:    "full",
        Aliases: []string{"f"},
        Usage:   "Fetch llms-full.txt instead of llms.txt",
    },
    &cli.BoolFlag{
        Name:    "verbose",
        Aliases: []string{"v"},
        Usage:   "Enable verbose/debug output",
        Sources: cli.EnvVars("LLXT_VERBOSE"),
    },
    &cli.DurationFlag{
        Name:    "timeout",
        Usage:   "HTTP request timeout",
        Value:   30 * time.Second,
        Sources: cli.EnvVars("LLXT_TIMEOUT"),
    },
    &cli.IntFlag{
        Name:    "retry",
        Usage:   "Number of retry attempts",
        Value:   3,
        Sources: cli.EnvVars("LLXT_RETRY"),
    },
},
```

### CLI Structure with urfave/cli v3

```go
import (
    "context"
    "os"

    "github.com/urfave/cli/v3"
)

func main() {
    cmd := &cli.Command{
        Name:    "llxt",
        Usage:   "Fetch or generate llms.txt files",
        Version: version, // Set by goreleaser
        Commands: []*cli.Command{
            fetchCommand(),
            addCommand(),
            listCommand(),
            searchCommand(),
            removeCommand(),
            registryCommand(), // For maintainers
        },
        Flags: globalFlags(),
        Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
            // Initialize logger based on verbose flag
            initLogger(cmd.Bool("verbose"))

            // Load configuration
            if err := loadConfig(cmd.String("config")); err != nil {
                return nil, err
            }

            return ctx, nil
        },
    }

    if err := cmd.Run(context.Background(), os.Args); err != nil {
        os.Exit(1)
    }
}

func fetchCommand() *cli.Command {
    return &cli.Command{
        Name:      "fetch",
        Usage:     "Fetch llms.txt for a tool/framework",
        ArgsUsage: "<name>",
        Flags: []cli.Flag{
            &cli.BoolFlag{Name: "full", Aliases: []string{"f"}, Usage: "Fetch llms-full.txt if available"},
        },
        Action: func(ctx context.Context, cmd *cli.Command) error {
            name := cmd.Args().First()
            if name == "" {
                return cli.Exit("name is required", ExitInvalidInput)
            }
            return fetchAction(ctx, cmd, name)
        },
    }
}

func registryCommand() *cli.Command {
    return &cli.Command{
        Name:  "registry",
        Usage: "Manage the bundled registry (for maintainers)",
        Commands: []*cli.Command{
            {
                Name:      "add",
                Usage:     "Add an entry to the bundled registry",
                ArgsUsage: "<key>",
                Flags: []cli.Flag{
                    &cli.StringFlag{Name: "url", Usage: "URL to llms.txt", Required: true},
                    &cli.StringFlag{Name: "full-url", Usage: "URL to llms-full.txt"},
                    &cli.StringFlag{Name: "name", Usage: "Display name", Required: true},
                },
                Action: registryAddAction,
            },
            {
                Name:      "remove",
                Usage:     "Remove an entry from the bundled registry",
                ArgsUsage: "<key>",
                Action:    registryRemoveAction,
            },
            {
                Name:   "validate",
                Usage:  "Validate all registry URLs are accessible",
                Action: registryValidateAction,
            },
            {
                Name:   "scrape",
                Usage:  "Scrape and update registry from directories",
                Action: registryScrapeAction,
            },
        },
    }
}
```

---

## Configuration Loading

### Koanf with CLI Flag Integration

```go
import (
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/providers/confmap"
    "github.com/knadh/koanf/providers/cliflagv3"
    toml "github.com/knadh/koanf/parsers/toml/v2"
)

var k = koanf.New(".")

type Config struct {
    HTTP     HTTPConfig               `koanf:"http"`
    Registry map[string]RegistryEntry `koanf:"registry"`
    Sources  map[string]GitHubSource  `koanf:"sources"`
}

func loadConfig(path string) error {
    // Load defaults
    if err := k.Load(confmap.Provider(map[string]any{
        "http.timeout":            "30s",
        "http.retry_count":        3,
        "http.retry_wait_time":    "100ms",
        "http.retry_max_wait_time": "2s",
    }, "."), nil); err != nil {
        return configErr.Wrapf(err, "failed to load defaults")
    }

    // Expand ~ in path
    path = expandPath(path)

    // Load config file if exists
    if _, err := os.Stat(path); err == nil {
        if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
            return configErr.
                Code(CodeConfigParse).
                With("path", path).
                Wrapf(err, "failed to parse config file")
        }
    }

    return nil
}

// Merge CLI flags into config
func mergeFlags(cmd *cli.Command) error {
    return k.Load(cliflagv3.Provider(cmd, "."), nil)
}
```

---

## Output Formatting

### AI-Friendly Default Output

```go
// Default: one key per line (easily parsed by AI agents)
func renderSourceListPlain(sources []Source, w io.Writer) {
    for _, s := range sources {
        fmt.Fprintln(w, s.Key)
    }
}

// Key-value format for info command
func renderInfoPlain(entry *RegistryEntry, w io.Writer) {
    fmt.Fprintf(w, "key=%s\n", entry.Key)
    fmt.Fprintf(w, "name=%s\n", entry.Name)
    fmt.Fprintf(w, "url=%s\n", entry.LLMsURL)
    if entry.LLMsFullURL != nil {
        fmt.Fprintf(w, "full_url=%s\n", *entry.LLMsFullURL)
    }
    if entry.Category != "" {
        fmt.Fprintf(w, "category=%s\n", entry.Category)
    }
    if entry.Description != "" {
        fmt.Fprintf(w, "description=%s\n", entry.Description)
    }
}
```

### Human-Friendly Output (--table flag)

```go
import (
    "os"
    "github.com/jedib0t/go-pretty/v6/table"
    "github.com/jedib0t/go-pretty/v6/text"
)

func renderSourceListTable(sources []Source) {
    t := table.NewWriter()
    t.SetOutputMirror(os.Stdout)
    t.SetStyle(table.StyleRounded)

    t.AppendHeader(table.Row{"Name", "Key", "Category", "URL"})

    for _, s := range sources {
        t.AppendRow(table.Row{
            s.DisplayName,
            s.Key,
            s.Category,
            text.WrapSoft(s.URL, 40),
        })
    }

    t.AppendFooter(table.Row{"", "", "Total", len(sources)})
    t.Render()
}
```

### Progress Spinners (stderr, not stdout)

Progress indicators are written to stderr to keep stdout clean for piping to AI agents.

```go
import (
    "os"
    "time"
    "github.com/jedib0t/go-pretty/v6/progress"
)

// Single operation spinner (fetch, generate)
func withSpinner(message string, fn func() error) error {
    pw := progress.NewWriter()
    pw.SetOutputWriter(os.Stderr) // IMPORTANT: stderr, not stdout
    pw.SetAutoStop(true)
    pw.SetTrackerLength(25)
    pw.SetStyle(progress.StyleDefault)
    pw.Style().Visibility.ETA = false
    pw.Style().Visibility.Percentage = false

    go pw.Render()

    tracker := &progress.Tracker{
        Message: message,
        Total:   0, // Indeterminate
    }
    pw.AppendTracker(tracker)

    err := fn()

    if err != nil {
        tracker.MarkAsErrored()
    } else {
        tracker.MarkAsDone()
    }

    // Wait for render to complete
    time.Sleep(100 * time.Millisecond)
    pw.Stop()

    return err
}

// Multiple operations with progress (validate, scrape)
func withProgressBar(message string, total int64, fn func(tracker *progress.Tracker) error) error {
    pw := progress.NewWriter()
    pw.SetOutputWriter(os.Stderr)
    pw.SetAutoStop(true)
    pw.SetTrackerLength(25)

    go pw.Render()

    tracker := &progress.Tracker{
        Message: message,
        Total:   total,
    }
    pw.AppendTracker(tracker)

    err := fn(tracker)

    if err != nil {
        tracker.MarkAsErrored()
    } else {
        tracker.MarkAsDone()
    }

    time.Sleep(100 * time.Millisecond)
    pw.Stop()

    return err
}

// Usage example
func fetchAction(ctx context.Context, cmd *cli.Command, name string) error {
    var content string

    err := withSpinner(fmt.Sprintf("Fetching %s...", name), func() error {
        var err error
        content, err = fetchLLMsTxt(name, cmd.Bool("full"))
        return err
    })

    if err != nil {
        return err
    }

    // Output raw content to stdout (for AI agents to consume)
    fmt.Print(content)
    return nil
}
```

### Validation with Progress

```go
func validateRegistry() error {
    entries := getAllEntries()
    total := int64(len(entries))

    var results []ValidationResult
    var mu sync.Mutex

    return withProgressBar("Validating registry URLs", total, func(tracker *progress.Tracker) error {
        var wg sync.WaitGroup
        sem := make(chan struct{}, 10) // Limit concurrency

        for _, entry := range entries {
            wg.Add(1)
            go func(e RegistryEntry) {
                defer wg.Done()
                sem <- struct{}{}
                defer func() { <-sem }()

                start := time.Now()
                err := validateURL(e.LLMsURL)
                duration := time.Since(start)

                mu.Lock()
                results = append(results, ValidationResult{
                    Key:      e.Key,
                    URL:      e.LLMsURL,
                    Error:    err,
                    Duration: duration,
                })
                mu.Unlock()

                tracker.Increment(1)
            }(entry)
        }

        wg.Wait()
        return nil
    })
}
```

---

## Project Structure

```
llxt/
├── cmd/
│   └── llxt/
│       └── main.go              # CLI entry point (urfave/cli)
├── internal/
│   ├── config/
│   │   ├── config.go            # Koanf config loading
│   │   └── types.go             # Config structs
│   ├── errors/
│   │   ├── codes.go             # Error codes
│   │   ├── domains.go           # Error domains
│   │   └── errors.go            # Error builders with oops
│   ├── registry/
│   │   ├── registry.go          # Registry lookup logic
│   │   ├── embed.go             # go:embed directives & loading
│   │   ├── entry.go             # RegistryEntry type & methods
│   │   ├── manage.go            # Registry management (add/remove)
│   │   └── validate.go          # URL validation
│   ├── http/
│   │   ├── client.go            # Resty client setup
│   │   └── fetch.go             # Fetch llms.txt content
│   ├── github/
│   │   ├── client.go            # GitHub API client (Resty)
│   │   └── trees.go             # GitHub Trees API
│   ├── generator/
│   │   └── generator.go         # llms.txt generation
│   ├── scraper/
│   │   ├── scraper.go           # Main scraper logic
│   │   └── parsers.go           # Site-specific parsers
│   └── ui/
│       ├── output.go            # Output formatting (plain, table, json)
│       └── progress.go          # Progress spinners (go-pretty)
├── registry/
│   ├── directory_entries.json   # Scraped from directory.llmstxt.cloud (embedded)
│   └── websites.json            # Scraped from llmstxt.site (embedded)
├── Taskfile.yml                 # Task runner configuration
├── .goreleaser.yaml             # Release configuration
├── go.mod
├── go.sum
├── README.md
└── PLAN.md
```

### Embedding Registry Data

```go
// internal/registry/embed.go
package registry

import (
    "embed"
    "encoding/json"
)

//go:embed ../../registry/directory_entries.json
var directoryEntriesJSON []byte

//go:embed ../../registry/websites.json
var websitesJSON []byte

// DirectoryEntry matches directory_entries.json format
type DirectoryEntry struct {
    Key         string  `json:"key"`
    Name        string  `json:"name"`
    Domain      string  `json:"domain"`
    LLMsURL     string  `json:"llms_url"`
    LLMsFullURL *string `json:"llms_full_url"`
}

// WebsiteEntry matches websites.json format
type WebsiteEntry struct {
    Name          string  `json:"name"`
    Domain        string  `json:"domain"`
    Description   string  `json:"description"`
    LLMsTxtURL    string  `json:"llmsTxtUrl"`
    LLMsFullTxtURL *string `json:"llmsFullTxtUrl,omitempty"`
    Category      string  `json:"category"`
    Favicon       string  `json:"favicon"`
    PublishedAt   string  `json:"publishedAt"`
}

var (
    directoryEntries []DirectoryEntry
    websiteEntries   []WebsiteEntry
    unified          map[string]*RegistryEntry // key -> entry
)

func init() {
    // Parse embedded JSON
    json.Unmarshal(directoryEntriesJSON, &directoryEntries)
    json.Unmarshal(websitesJSON, &websiteEntries)

    // Build unified registry (websites.json has richer data, use it when available)
    unified = make(map[string]*RegistryEntry)

    // First, load directory entries
    for _, e := range directoryEntries {
        unified[e.Key] = &RegistryEntry{
            Key:         e.Key,
            Name:        e.Name,
            Domain:      e.Domain,
            LLMsURL:     e.LLMsURL,
            LLMsFullURL: e.LLMsFullURL,
        }
    }

    // Then, merge/override with website entries (richer data)
    for _, e := range websiteEntries {
        key := generateKey(e.Name) // slugify
        if existing, ok := unified[key]; ok {
            // Enrich existing entry
            existing.Description = e.Description
            existing.Category = e.Category
        } else {
            unified[key] = &RegistryEntry{
                Key:         key,
                Name:        e.Name,
                Domain:      e.Domain,
                Description: e.Description,
                Category:    e.Category,
                LLMsURL:     e.LLMsTxtURL,
                LLMsFullURL: e.LLMsFullTxtURL,
            }
        }
    }
}
```

---

## Taskfile Configuration

```yaml
# https://taskfile.dev

version: '3'

vars:
  BINARY_NAME: llxt
  BUILD_DIR: ./dist
  VERSION:
    sh: git describe --tags --always --dirty 2>/dev/null || echo "dev"
  COMMIT:
    sh: git rev-parse --short HEAD 2>/dev/null || echo "unknown"
  BUILD_TIME:
    sh: date -u '+%Y-%m-%dT%H:%M:%SZ'

env:
  CGO_ENABLED: '0'

tasks:
  default:
    desc: Show available tasks
    cmds:
      - task --list

  # Development
  dev:
    desc: Run the application in development mode
    cmds:
      - go run ./cmd/llxt {{.CLI_ARGS}}

  # Building
  build:
    desc: Build the binary
    cmds:
      - go build -ldflags "-s -w -X main.version={{.VERSION}} -X main.commit={{.COMMIT}} -X main.buildTime={{.BUILD_TIME}}" -o {{.BUILD_DIR}}/{{.BINARY_NAME}} ./cmd/llxt
    sources:
      - ./**/*.go
      - go.mod
      - go.sum
    generates:
      - "{{.BUILD_DIR}}/{{.BINARY_NAME}}"

  build:all:
    desc: Build for all platforms
    cmds:
      - task: build:linux
      - task: build:darwin
      - task: build:windows

  build:linux:
    desc: Build for Linux
    env:
      GOOS: linux
      GOARCH: amd64
    cmds:
      - go build -ldflags "-s -w" -o {{.BUILD_DIR}}/{{.BINARY_NAME}}-linux-amd64 ./cmd/llxt

  build:darwin:
    desc: Build for macOS
    env:
      GOOS: darwin
      GOARCH: arm64
    cmds:
      - go build -ldflags "-s -w" -o {{.BUILD_DIR}}/{{.BINARY_NAME}}-darwin-arm64 ./cmd/llxt

  build:windows:
    desc: Build for Windows
    env:
      GOOS: windows
      GOARCH: amd64
    cmds:
      - go build -ldflags "-s -w" -o {{.BUILD_DIR}}/{{.BINARY_NAME}}-windows-amd64.exe ./cmd/llxt

  # Testing
  test:
    desc: Run tests
    cmds:
      - go test -v -race -cover ./...

  test:coverage:
    desc: Run tests with coverage report
    cmds:
      - go test -v -race -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out -o coverage.html
      - echo "Coverage report: coverage.html"

  # Linting
  lint:
    desc: Run linters
    cmds:
      - golangci-lint run ./...

  lint:fix:
    desc: Run linters and fix issues
    cmds:
      - golangci-lint run --fix ./...

  # Formatting
  fmt:
    desc: Format code
    cmds:
      - go fmt ./...
      - goimports -w .

  # Dependencies
  deps:
    desc: Download dependencies
    cmds:
      - go mod download

  deps:tidy:
    desc: Tidy go.mod
    cmds:
      - go mod tidy

  deps:upgrade:
    desc: Upgrade all dependencies
    cmds:
      - go get -u ./...
      - go mod tidy

  # Registry Management
  registry:validate:
    desc: Validate all registry URLs
    cmds:
      - go run ./cmd/llxt registry validate

  registry:scrape:
    desc: Scrape and update registry from directories
    cmds:
      - go run ./cmd/llxt registry scrape

  registry:add:
    desc: Add an entry to the registry
    cmds:
      - go run ./cmd/llxt registry add {{.CLI_ARGS}}

  # Code Generation
  generate:
    desc: Run go generate
    cmds:
      - go generate ./...

  # Cleaning
  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf {{.BUILD_DIR}}
      - rm -f coverage.out coverage.html

  # Release
  release:dry-run:
    desc: Test release process without publishing
    cmds:
      - goreleaser release --snapshot --clean

  release:
    desc: Create a new release
    cmds:
      - goreleaser release --clean

  # Installation
  install:
    desc: Install the binary locally
    cmds:
      - go install -ldflags "-s -w -X main.version={{.VERSION}}" ./cmd/llxt

  # All-in-one
  all:
    desc: Run fmt, lint, test, and build
    cmds:
      - task: fmt
      - task: lint
      - task: test
      - task: build
```

---

## GoReleaser Configuration

```yaml
# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

project_name: llxt

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - main: ./cmd/llxt
    binary: llxt
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.buildTime={{.Date}}

archives:
  - formats: [tar.gz]
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
    format_overrides:
      - goos: windows
        formats: [zip]
    files:
      - README.md
      - LICENSE

checksum:
  name_template: 'checksums.txt'

snapshot:
  version_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
      - "^chore:"
  groups:
    - title: Features
      regexp: "^.*feat[(\\w)]*:+.*$"
      order: 0
    - title: Bug Fixes
      regexp: "^.*fix[(\\w)]*:+.*$"
      order: 1
    - title: Others
      order: 999

brews:
  - name: llxt
    homepage: https://github.com/yourusername/llxt
    description: Fetch or generate llms.txt files
    license: MIT
    repository:
      owner: yourusername
      name: homebrew-tap
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com

release:
  footer: >-
    ---

    **Full Changelog**: https://github.com/yourusername/llxt/compare/{{ .PreviousTag }}...{{ .Tag }}
```

---

## Implementation Phases

### Phase 1: Core MVP
- [ ] Project skeleton with go.mod and dependencies
- [ ] Error handling infrastructure with oops
- [ ] CLI skeleton with `fetch` command (urfave/cli v3)
- [ ] **Embed existing scraped data** (directory_entries.json, websites.json)
- [ ] Unified registry loading from both JSON sources
- [ ] HTTP client with retry & circuit breaker (resty v3)
- [ ] Fetch llms.txt/llms-full.txt from registry URLs
- [ ] **AI-friendly output** (raw content to stdout)
- [ ] **Progress spinners** to stderr (go-pretty/progress)
- [ ] Taskfile setup with build/test/lint tasks

### Phase 2: List & Search
- [ ] `list` command with AI-friendly default (one key per line)
- [ ] `--table` flag for human-friendly table output
- [ ] `--json` flag for JSON output
- [ ] `--category` filter
- [ ] `search` command with fuzzy matching
- [ ] `info` command for source details

### Phase 3: User Config
- [ ] Config file loading with koanf (TOML)
- [ ] `add` command for URL-based sources
- [ ] `remove` command
- [ ] Config file creation if not exists
- [ ] Merge user sources with bundled registry

### Phase 4: GitHub Generation
- [ ] GitHub Trees API integration
- [ ] `add` command for GitHub-based sources
- [ ] On-the-fly llms.txt generation with progress
- [ ] Glob pattern support with exclude

### Phase 5: Registry Management (Maintainers)
- [ ] `registry add` command for bundled registry
- [ ] `registry remove` command
- [ ] `registry validate` command with concurrent validation & progress
- [ ] `registry scrape` command (refresh from directories)

### Phase 6: Polish
- [ ] Verbose/debug output with slog
- [ ] Error messages with suggestions & hints
- [ ] Shell completions (bash, zsh, fish, powershell)
- [ ] `--quiet` flag to suppress progress spinners

### Phase 7: Release
- [ ] GoReleaser configuration
- [ ] Homebrew tap setup
- [ ] GitHub Actions CI/CD
- [ ] README documentation

---

## Component Breakdown

| Component                    | Library     | Effort   | Description                           |
|------------------------------|-------------|----------|---------------------------------------|
| CLI structure                | urfave/cli  | Trivial  | Commands, flags, help                 |
| Error infrastructure         | oops        | Easy     | Domains, codes, builders              |
| **Embed scraped data**       | embed       | Trivial  | Load directory_entries + websites     |
| Registry unification         | stdlib      | Easy     | Merge two JSON formats into one       |
| Config loading               | koanf       | Easy     | Load ~/.config/llxt/sources.toml      |
| HTTP client                  | resty       | Easy     | Retry, circuit breaker, timeouts      |
| Registry fetch               | resty       | Trivial  | HTTP GET llms.txt URLs                |
| **Progress spinners**        | go-pretty   | Easy     | Indeterminate & determinate progress  |
| **AI-friendly output**       | stdlib      | Trivial  | Plain text, key=value formats         |
| Human-friendly output        | go-pretty   | Trivial  | Tables with --table flag              |
| JSON output                  | stdlib      | Trivial  | --json flag                           |
| Fuzzy search                 | stdlib      | Easy     | Substring/prefix matching             |
| GitHub Trees API             | resty       | Easy     | List files in repo                    |
| llms.txt generation          | stdlib      | Easy     | Format results with progress          |
| Logging                      | slog        | Trivial  | Structured logging                    |
| Registry management          | stdlib      | Easy     | Add/remove/validate entries           |
| Concurrent validation        | stdlib      | Easy     | Parallel URL checks with semaphore    |
| Registry scraper             | goquery     | Medium   | HTML parsing for registry sites       |

**Estimated Total:** 1000-1500 lines of Go

---

## Pre-Implementation Setup

**Move existing scraped data to the registry directory:**

```bash
mkdir -p registry
mv directory_entries.json registry/
mv websites.json registry/
```

**Note:** The scraped data files are currently in the project root. They need to be moved to `registry/` for the go:embed directives to work correctly.

---

## Open Questions

1. **Registry key format?** — Should we use slugs (`vercel-ai`) or allow spaces (`Vercel AI SDK`)?
   - **Recommendation**: Use slugs for keys, display names for output

2. **Caching?** — Should fetched llms.txt content be cached locally?
   - **Recommendation**: Optional caching with `--cache` flag, default no cache

3. **Full by default?** — Should `--full` be the default behavior when available?
   - **Recommendation**: No, keep `--full` as opt-in to reduce bandwidth

4. **Scraper library?** — Use colly, goquery, or just stdlib html parsing?
   - **Recommendation**: Start with goquery for simplicity

5. **GitHub rate limiting?** — How to handle unauthenticated API limits (60 req/hr)?
   - **Recommendation**: Support `GITHUB_TOKEN` env var for higher limits

---

## References

### External
- [llms.txt specification](https://llmstxt.org/)
- [llmstxt.cloud directory](https://directory.llmstxt.cloud/)
- [llmstxt.site](https://llmstxt.site/)
- [GitHub Trees API](https://docs.github.com/en/rest/git/trees)

### Local Documentation (`.docs/`)
- **koanf** — [.docs/koanf.md](.docs/koanf.md)
- **urfave/cli** — [.docs/urfave_cli/](.docs/urfave_cli/)
- **resty** — [.docs/resty/](.docs/resty/)
- **go-pretty** — [.docs/go_pretty/](.docs/go_pretty/)
- **oops** — [.docs/OOPS.md](.docs/OOPS.md)
- **goreleaser** — [.docs/goreleaser/](.docs/goreleaser/)
