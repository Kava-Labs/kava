package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/migrate"
)

// kvd custom flags
const (
	flagInvCheckPeriod       = "inv-check-period"
	flagMempoolEnableAuth    = "mempool.enable-authentication"
	flagMempoolAuthAddresses = "mempool.authorized-addresses"
)

var invCheckPeriod uint

func main() {
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	app.SetBip44CoinType(config)
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "kvd",
		Short:             "Kava Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(
		genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, app.DefaultNodeHome),
		migrate.MigrateGenesisCmd(ctx, cdc),
		migrate.ValidateGenesisInitCmd(ctx, cdc),
		genutilcli.GenTxCmd(
			ctx,
			cdc,
			app.ModuleBasics,
			staking.AppModuleBasic{},
			auth.GenesisAccountIterator{},
			app.DefaultNodeHome,
			app.DefaultCLIHome),
		genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics),
		AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
		testnetCmd(ctx, cdc, app.ModuleBasics, auth.GenesisAccountIterator{}),
		flags.NewCompletionCmd(rootCmd, true),
	)

	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "KA", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	startCmd, _, err := rootCmd.Find([]string{"start"})
	if err != nil {
		panic(fmt.Sprintf("could not find 'start' command on root command: %s", err))
	}
	startCmd.Flags().Bool(flagMempoolEnableAuth, false, "Configure the mempool to only accept transactions from authorized addresses")
	err = viper.BindPFlag(flagMempoolEnableAuth, startCmd.Flags().Lookup(flagMempoolEnableAuth))
	if err != nil {
		panic(fmt.Sprintf("failed to bind flag: %s", err))
	}
	startCmd.Flags().StringSlice(flagMempoolAuthAddresses, []string{}, "Additional addresses to accept transactions from when the mempool is running in authorized mode (comma separated kava addresses)")
	err = viper.BindPFlag(flagMempoolAuthAddresses, startCmd.Flags().Lookup(flagMempoolAuthAddresses))
	if err != nil {
		panic(fmt.Sprintf("failed to bind flag: %s", err))
	}

	// run main command
	err = executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	var cache sdk.MultiStorePersistentCache

	if viper.GetBool(server.FlagInterBlockCache) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range viper.GetIntSlice(server.FlagUnsafeSkipUpgrades) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags()
	if err != nil {
		panic(err)
	}

	mempoolEnableAuth := viper.GetBool(flagMempoolEnableAuth)
	mempoolAuthAddresses, err := accAddressesFromBech32(viper.GetStringSlice(flagMempoolAuthAddresses)...)
	if err != nil {
		panic(fmt.Sprintf("could not get authorized address from config: %v", err))
	}

	return app.NewApp(
		logger, db, traceStore,
		app.AppOptions{
			SkipLoadLatest:       false,
			SkipUpgradeHeights:   skipUpgradeHeights,
			InvariantCheckPeriod: invCheckPeriod,
			MempoolEnableAuth:    mempoolEnableAuth,
			MempoolAuthAddresses: mempoolAuthAddresses,
		},
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(viper.GetUint64(server.FlagHaltHeight)),
		baseapp.SetHaltTime(viper.GetUint64(server.FlagHaltTime)),
		baseapp.SetInterBlockCache(cache),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		opts := app.AppOptions{
			SkipLoadLatest:       true,
			InvariantCheckPeriod: uint(1),
		}
		tempApp := app.NewApp(logger, db, traceStore, opts)
		err := tempApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return tempApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	opts := app.AppOptions{
		SkipLoadLatest:       false,
		InvariantCheckPeriod: uint(1),
	}
	tempApp := app.NewApp(logger, db, traceStore, opts)
	return tempApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

func accAddressesFromBech32(addresses ...string) ([]sdk.AccAddress, error) {
	var decodedAddresses []sdk.AccAddress
	for _, s := range addresses {
		a, err := sdk.AccAddressFromBech32(s)
		if err != nil {
			return nil, err
		}
		decodedAddresses = append(decodedAddresses, a)
	}
	return decodedAddresses, nil
}
