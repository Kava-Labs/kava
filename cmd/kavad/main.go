package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/kava-labs/kava/internal/app"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "kavad",
		Short:             "Kava Daemon",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, server.DefaultAppInit,
		server.ConstructAppCreator(newApp, "kava"),
		server.ConstructAppExporter(exportAppState, "kava"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.kavad")
	executor := cli.PrepareBaseCmd(rootCmd, "KV", rootDir)
	executor.Execute()
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewBasecoinApp(logger, db)
}

func exportAppState(logger log.Logger, db dbm.DB) (json.RawMessage, error) {
	bapp := app.NewBasecoinApp(logger, db)
	return bapp.ExportAppStateJSON()
}
