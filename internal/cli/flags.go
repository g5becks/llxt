package cli

import (
	"time"

	"github.com/urfave/cli/v3"
)

const defaultTimeout = 30 * time.Second

// GlobalFlags returns the global flags for the CLI.
func GlobalFlags() []cli.Flag {
	return []cli.Flag{
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
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "Enable verbose/debug output",
			Sources: cli.EnvVars("LLXT_VERBOSE"),
		},
		&cli.DurationFlag{
			Name:    "timeout",
			Usage:   "HTTP request timeout",
			Value:   defaultTimeout,
			Sources: cli.EnvVars("LLXT_TIMEOUT"),
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "Suppress progress spinners",
		},
	}
}
