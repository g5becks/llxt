package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	llxtcli "github.com/g5becks/llxt/internal/cli"
	"github.com/g5becks/llxt/internal/cli/commands"
)

//nolint:gochecknoglobals // version info set by build flags
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	cmd := &cli.Command{
		Name:    "llxt",
		Usage:   "Fetch or generate llms.txt files for AI agents",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildTime),
		Flags:   llxtcli.GlobalFlags(),
		Commands: []*cli.Command{
			commands.FetchCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(llxtcli.ExitGeneralError)
	}
}
