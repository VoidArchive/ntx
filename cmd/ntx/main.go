package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

var cmd struct {
	Version struct{} `cmd:"" help:"Show version information"`
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
	default:
		fmt.Println("ntx - NEPSE Stock Aggregator")
		fmt.Println()
		fmt.Println("Commands coming soon:")
		fmt.Println("  ntx market      # Market overview")
		fmt.Println("  ntx price NABIL # Get stock price")
		fmt.Println("  ntx analyze     # Stock analysis")
		fmt.Println()
		fmt.Println("Run 'ntx --help' for more information.")
		os.Exit(0)
	}
}
