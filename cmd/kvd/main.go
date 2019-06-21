package main

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	initPkg "github.com/kava-labs/kava/init"
)

// kvd custom flags
const flagInvCheckPeriod = "inv-check-period"

var invCheckPeriod uint

func main() {
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.Bech32MainPrefix, app.Bech32MainPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator, app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, app.Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "kvd",
		Short:             "Kava Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(initPkg.InitCmd(ctx, cdc))
	rootCmd.AddCommand(initPkg.CollectGenTxsCmd(ctx, cdc))
	rootCmd.AddCommand(initPkg.TestnetFilesCmd(ctx, cdc))
	rootCmd.AddCommand(initPkg.GenTxCmd(ctx, cdc))
	rootCmd.AddCommand(initPkg.AddGenesisAccountCmd(ctx, cdc))
	rootCmd.AddCommand(initPkg.ValidateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(client.NewCompletionCmd(rootCmd, true))

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "GA", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewApp(
		logger, db, traceStore, true, invCheckPeriod,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		gApp := app.NewApp(logger, db, traceStore, false, uint(1))
		err := gApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return gApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	gApp := app.NewApp(logger, db, traceStore, true, uint(1))
	return gApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
