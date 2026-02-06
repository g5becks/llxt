// Package commands provides CLI command implementations.
package commands

import (
	"context"
	"fmt"

	"github.com/samber/oops"
	"github.com/urfave/cli/v3"

	llxtcli "github.com/g5becks/llxt/internal/cli"
	httpclient "github.com/g5becks/llxt/internal/http"
	"github.com/g5becks/llxt/internal/registry"
	"github.com/g5becks/llxt/internal/ui"
)

// FetchCommand returns the fetch command.
func FetchCommand() *cli.Command {
	return &cli.Command{
		Name:      "fetch",
		Usage:     "Fetch llms.txt for a tool/framework",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "full",
				Aliases: []string{"f"},
				Usage:   "Fetch llms-full.txt if available",
			},
		},
		Action: fetchAction,
	}
}

func fetchAction(_ context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	if name == "" {
		return cli.Exit("name is required\n\nUsage: llxt fetch <name>", llxtcli.ExitInvalidInput)
	}

	full := cmd.Bool("full")
	quiet := cmd.Bool("quiet")
	verbose := cmd.Bool("verbose")

	// Lookup in registry
	entry, err := registry.Lookup(name)
	if err != nil {
		msg := oops.GetPublic(err, fmt.Sprintf("Source %q not found", name))
		return cli.Exit(msg, llxtcli.ExitNotFound)
	}

	// Create HTTP client
	cfg := httpclient.DefaultConfig()
	cfg.Verbose = verbose
	fetcher := httpclient.NewFetcher(cfg)
	defer fetcher.Close()

	var content string

	fetchFn := func() error {
		var fetchErr error
		content, fetchErr = fetcher.FetchLLMsTxt(entry.LLMsURL, entry.LLMsFullURL, full)
		return fetchErr
	}

	if quiet {
		err = fetchFn()
	} else {
		err = ui.WithSpinner(fmt.Sprintf("Fetching %s...", name), fetchFn)
	}

	if err != nil {
		msg := oops.GetPublic(err, "Failed to fetch llms.txt")
		return cli.Exit(msg, llxtcli.ExitNetworkError)
	}

	// Output raw content to stdout (AI-friendly)
	//nolint:forbidigo // stdout output is intentional for AI consumption
	fmt.Print(content)
	return nil
}
