package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kong"

	"github.com/voidarchive/ntx/cmd/ntx/cli"
	"github.com/voidarchive/ntx/cmd/ntx/tui"
	"github.com/voidarchive/ntx/internal/database"
	"github.com/voidarchive/ntx/internal/portfolio"
)

var cmd struct {
	Import struct {
		File string `arg:"" help:"Path to Meroshare CSV file" type:"path"`
	} `cmd:"" help:"Import transactions from Meroshare CSV file"`

	ImportWacc struct {
		File string `arg:"" help:"Path to Meroshare WACC Report CSV file" type:"path"`
	} `cmd:"" name:"import-wacc" help:"Import cost data from Meroshare WACC Report"`

	Holdings struct{} `cmd:"" help:"List all holdings"`

	Summary struct{} `cmd:"" help:"Show portfolio summary"`

	Transactions struct {
		Symbol string `short:"s" help:"Filter by symbol"`
		Type   string `short:"t" help:"Filter by type" enum:",buy,sell,ipo,bonus,rights,merger_in,merger_out,demat,rearrangement" default:""`
		Limit  int    `short:"l" default:"10" help:"Limit results"`
		Offset int    `short:"o" default:"0" help:"Offset for pagination"`
	} `cmd:"" help:"List transactions"`

	Sync struct{} `cmd:"" help:"Fetch latest prices and update holdings"`
}

func main() {
	// Check if no args - launch TUI directly
	if len(os.Args) == 1 {
		db, err := database.OpenDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		if err := database.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}

		service := portfolio.NewService(db)
		if err := tui.Run(service); err != nil {
			log.Fatal(err)
		}
		return
	}

	ctx := kong.Parse(&cmd)

	db, err := database.OpenDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	service := portfolio.NewService(db)

	switch ctx.Command() {
	case "import <file>":
		if cmd.Import.File == "" {
			ctx.FatalIfErrorf(fmt.Errorf("file path required"))
		}
		cli.Import(service, cmd.Import.File)

	case "import-wacc <file>":
		if cmd.ImportWacc.File == "" {
			ctx.FatalIfErrorf(fmt.Errorf("file path required"))
		}
		cli.ImportWacc(service, cmd.ImportWacc.File)

	case "holdings":
		cli.Holdings(service)

	case "summary":
		cli.Summary(service)

	case "transactions":
		cli.Transactions(
			service,
			cmd.Transactions.Symbol,
			cmd.Transactions.Type,
			cmd.Transactions.Limit,
			cmd.Transactions.Offset,
		)

	case "sync":
		cli.Sync(service)

	default:
		// No command given - launch TUI
		if err := tui.Run(service); err != nil {
			log.Fatal(err)
		}
	}
}
