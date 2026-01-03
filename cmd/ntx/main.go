package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var cmd struct {
	Version  struct{}    `cmd:"" help:"Show version information"`
	Backfill BackfillCmd `cmd:"" help:"Backfill historical price data (1 year)"`
}

func main() {
	ctx := kong.Parse(&cmd,
		kong.Name("ntx"),
		kong.Description("NEPSE Stock Aggregator - Market data, analysis, and insights"),
		kong.UsageOnError(),
	)

	switch ctx.Command() {
	case "version":
		fmt.Println("ntx v0.1.0")
	case "backfill":
		if err := cmd.Backfill.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Println("ntx - NEPSE Stock Aggregator")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  ntx backfill    # Fetch 1 year of historical prices")
		fmt.Println("  ntx market      # Market overview (coming soon)")
		fmt.Println("  ntx price NABIL # Get stock price (coming soon)")
		fmt.Println()
		fmt.Println("Run 'ntx --help' for more information.")
		os.Exit(0)
	}
}
