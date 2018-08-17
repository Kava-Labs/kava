package main

import (
	"encoding/json"
	"io"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/kava-labs/kava/internal/app"
)

func main() {
	// Create an app codec
	cdc := app.CreateKavaAppCodec()

	// Create a server context (a struct of a tendermint config and a logger)
	ctx := server.NewDefaultContext()

	// Create the root kvd command
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "kvd",
		Short:             "Kava Daemon",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	// Add server commands to kvd, passing in the app
	appInit := app.KavaAppInit()
	appCreator := server.ConstructAppCreator(newApp, "kava") // init db before calling newApp
	appExporter := server.ConstructAppExporter(exportAppStateAndTMValidators, "kava")

	server.AddCommands(ctx, cdc, rootCmd, appInit, appCreator, appExporter)

	// handle envs and add some flags and stuff
	executor := cli.PrepareBaseCmd(rootCmd, "KV", app.DefaultNodeHome)

	// Run kvd
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewKavaApp(logger, db, traceStore, baseapp.SetPruning(viper.GetString("pruning")))
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB, traceStore io.Writer) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	tempApp := app.NewKavaApp(logger, db, traceStore)
	return tempApp.ExportAppStateAndValidators()
}
