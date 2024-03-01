package cmd

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	cometbftdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	ethermintflags "github.com/evmos/ethermint/server/flags"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
	metricstypes "github.com/kava-labs/kava/x/metrics/types"
)

const (
	flagMempoolEnableAuth    = "mempool.enable-authentication"
	flagMempoolAuthAddresses = "mempool.authorized-addresses"
	flagSkipLoadLatest       = "skip-load-latest"
)

// appCreator holds functions used by the sdk server to control the kava app.
// The methods implement types in cosmos-sdk/server/types
type appCreator struct {
	encodingConfig params.EncodingConfig
}

// newApp loads config from AppOptions and returns a new app.
func (ac appCreator) newApp(
	logger log.Logger,
	db cometbftdb.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache sdk.MultiStorePersistentCache
	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
	snapshotDir := filepath.Join(homeDir, "data", "snapshots") // TODO can these directory names be imported from somewhere?
	snapshotDB, err := cometbftdb.NewDB("metadata", server.GetAppDBBackend(appOpts), snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}

	mempoolEnableAuth := cast.ToBool(appOpts.Get(flagMempoolEnableAuth))
	mempoolAuthAddresses, err := accAddressesFromBech32(
		cast.ToStringSlice(appOpts.Get(flagMempoolAuthAddresses))...,
	)
	if err != nil {
		panic(fmt.Sprintf("could not get authorized address from config: %v", err))
	}

	iavlDisableFastNode := appOpts.Get(server.FlagDisableIAVLFastNode)
	if iavlDisableFastNode == nil {
		iavlDisableFastNode = true
	}

	snapshotOptions := snapshottypes.NewSnapshotOptions(
		cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval)),
		cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent)),
	)

	// Setup chainId
	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if len(chainID) == 0 {
		// fallback to genesis chain-id
		appGenesis, err := tmtypes.GenesisDocFromFile(filepath.Join(homeDir, "config", "genesis.json"))
		if err != nil {
			panic(err)
		}
		chainID = appGenesis.ChainID
	}

	skipLoadLatest := false
	if appOpts.Get(flagSkipLoadLatest) != nil {
		skipLoadLatest = cast.ToBool(appOpts.Get(flagSkipLoadLatest))
	}

	return app.NewApp(
		logger, db, homeDir, traceStore, ac.encodingConfig,
		app.Options{
			SkipLoadLatest:        skipLoadLatest,
			SkipUpgradeHeights:    skipUpgradeHeights,
			SkipGenesisInvariants: cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants)),
			InvariantCheckPeriod:  cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
			MempoolEnableAuth:     mempoolEnableAuth,
			MempoolAuthAddresses:  mempoolAuthAddresses,
			EVMTrace:              cast.ToString(appOpts.Get(ethermintflags.EVMTracer)),
			EVMMaxGasWanted:       cast.ToUint64(appOpts.Get(ethermintflags.EVMMaxTxGasWanted)),
			TelemetryOptions:      metricstypes.TelemetryOptionsFromAppOpts(appOpts),
		},
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(strings.Replace(cast.ToString(appOpts.Get(server.FlagMinGasPrices)), ";", ",", -1)),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))), // TODO what is this?
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetIAVLCacheSize(cast.ToInt(appOpts.Get(server.FlagIAVLCacheSize))),
		baseapp.SetIAVLDisableFastNode(cast.ToBool(iavlDisableFastNode)),
		baseapp.SetIAVLLazyLoading(cast.ToBool(appOpts.Get(server.FlagIAVLLazyLoading))),
		baseapp.SetChainID(chainID),
	)
}

// appExport writes out an app's state to json.
func (ac appCreator) appExport(
	logger log.Logger,
	db cometbftdb.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	options := app.DefaultOptions
	options.SkipLoadLatest = true
	options.InvariantCheckPeriod = cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod))

	var tempApp *app.App
	if height != -1 {
		tempApp = app.NewApp(logger, db, homePath, traceStore, ac.encodingConfig, options)

		if err := tempApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		tempApp = app.NewApp(logger, db, homePath, traceStore, ac.encodingConfig, options)
	}
	return tempApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// addStartCmdFlags adds flags to the server start command.
func (ac appCreator) addStartCmdFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

// accAddressesFromBech32 converts a slice of bech32 encoded addresses into a slice of address types.
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
