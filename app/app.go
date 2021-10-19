package app

import (
	"io"
	"os"

	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/kava-labs/kava/app/ante"
)

const (
	appName          = "kava"
	Bech32MainPrefix = "kava"
	Bip44CoinType    = 459 // see https://github.com/satoshilabs/slips/blob/master/slip-0044.md
)

var (
	// default home directories for expected binaries
	DefaultCLIHome  = os.ExpandEnv("$HOME/.kvcli") // TODO
	DefaultNodeHome = os.ExpandEnv("$HOME/.kvd")

	// ModuleBasics manages simple versions of full app modules. It's used for things such as codec registration and genesis file verification.
	ModuleBasics = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler,
			distr.ProposalHandler,
			upgradeclient.ProposalHandler, // TODO remove upgrade
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		vesting.AppModuleBasic{},
	)

	// module account permissions
	// if these are changed, then the permissions
	// must also be migrated during a chain upgrade
	mAccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		distrtypes.ModuleName: true,
	}
)

// Verify app interface at compile time
var _ simapp.App = (*App)(nil)

// AppOptions bundles several configuration params for an App.
// The zero value can be used as a sensible default.
type AppOptions struct {
	SkipLoadLatest       bool
	SkipUpgradeHeights   map[int64]bool
	InvariantCheckPeriod uint
	MempoolEnableAuth    bool
	MempoolAuthAddresses []sdk.AccAddress
}

// App represents an extended ABCI application
type App struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	// keepers from all the modules
	accountKeeper  auth.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	stakingKeeper  stakingkeeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	mintKeeper     mintkeeper.Keeper
	distrKeeper    distrkeeper.Keeper
	govKeeper      govkeeper.Keeper
	crisisKeeper   crisiskeeper.Keeper
	upgradeKeeper  upgradekeeper.Keeper
	paramsKeeper   paramskeeper.Keeper
	evidenceKeeper evidencekeeper.Keeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager
}

// NewApp returns a reference to an initialized App.
func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts AppOptions, baseAppOptions ...func(*bam.BaseApp)) *App {

	cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(
		bam.MainStoreKey, authtypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey, evidencetypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(params.TStoreKey)

	var app = &App{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: appOpts.InvariantCheckPeriod,
		keys:           keys,
		tkeys:          tkeys,
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(app.cdc, keys[paramstypes.StoreKey], tkeys[params.TStoreKey])
	authSubspace := app.paramsKeeper.Subspace(authtypes.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(banktypes.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(stakingtypes.DefaultParamspace)
	mintSubspace := app.paramsKeeper.Subspace(minttypes.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distrtypes.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashingtypes.DefaultParamspace)
	govSubspace := app.paramsKeeper.Subspace(govtypes.DefaultParamspace).WithKeyTable(gov.ParamKeyTable())
	evidenceSubspace := app.paramsKeeper.Subspace(evidencetypes.DefaultParamspace)
	crisisSubspace := app.paramsKeeper.Subspace(crisistypes.DefaultParamspace)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		keys[authtypes.StoreKey],
		authSubspace,
		authtypes.ProtoBaseAccount,
		maccPerms,
	)
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		bankSubspace,
		app.BlacklistedAccAddrs(),
	)
	stakingKeeper := staking.NewKeeper(
		app.cdc,
		keys[stakingtypes.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		stakingSubspace,
	)
	app.mintKeeper = mint.NewKeeper(
		app.cdc,
		keys[minttypes.StoreKey],
		mintSubspace,
		&stakingKeeper,
		app.accountKeeper,
		app.bankKeeper,
		authtypes.FeeCollectorName,
	)
	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		keys[distrtypes.StoreKey],
		distrSubspace,
		&stakingKeeper,
		app.accountKeeper,
		app.bankKeeper,
		authtypes.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)
	app.slashingKeeper = slashing.NewKeeper(
		app.cdc,
		keys[slashingtypes.StoreKey],
		&stakingKeeper,
		slashingSubspace,
	)
	app.crisisKeeper = crisis.NewKeeper(
		crisisSubspace,
		app.invCheckPeriod,
		app.bankKeeper,
		auth.FeeCollectorName,
	)
	app.upgradeKeeper = upgrade.NewKeeper(
		appOpts.SkipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		app.cdc,
	)

	// create evidence keeper with router
	evidenceKeeper := evidence.NewKeeper(
		app.cdc,
		keys[evidencetypes.StoreKey],
		evidenceSubspace,
		&app.stakingKeeper,
		app.slashingKeeper,
	)
	evidenceRouter := evidencetypes.NewRouter()
	evidenceKeeper.SetRouter(evidenceRouter)
	app.evidenceKeeper = *evidenceKeeper

	// create gov keeper with router
	govRouter := govtypes.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramstypes.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper))
	app.govKeeper = gov.NewKeeper(
		app.cdc,
		keys[govtypes.StoreKey],
		govSubspace,
		app.accountKeeper,
		app.bankKeeper,
		&stakingKeeper,
		govRouter,
	)

	// register the staking hooks
	// NOTE: These keepers are passed by reference above, so they will contain these hooks.
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	// create the module manager (Note: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.)
	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper),
		vesting.NewAppModule(app.accountKeeper, app.bankKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		gov.NewAppModule(app.govKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper),
		mint.NewAppModule(app.mintKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
		distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// Auction.BeginBlocker will close out expired auctions and pay debt back to cdp.
	// So it should be run before cdp.BeginBlocker which cancels out debt with stable and starts more auctions.
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName, minttypes.ModuleName, distrtypes.ModuleName, slashingtypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(crisistypes.ModuleName, govtypes.ModuleName, stakingtypes.ModuleName)

	app.mm.SetOrderInitGenesis(
		authtypes.ModuleName, // loads all accounts - should run before any module with a module account
		distrtypes.ModuleName,
		stakingtypes.ModuleName, banktypes.ModuleName, slashingtypes.ModuleName,
		govtypes.ModuleName, minttypes.ModuleName, evidencetypes.ModuleName,
		crisistypes.ModuleName,  // runs the invariants at genesis - should run after other modules
		genutiltypes.ModuleName, // genutils must occur after staking so that pools are properly initialized with tokens from genesis accounts.
	)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: This is not required for apps that don't use the simulator for fuzz testing
	// transactions.
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		gov.NewAppModule(app.govKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper),
		mint.NewAppModule(app.mintKeeper),
		distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	// initialize the app
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	var antehandler sdk.AnteHandler
	if appOpts.MempoolEnableAuth {
		var getAuthorizedAddresses ante.AddressFetcher = func(sdk.Context) []sdk.AccAddress { return appOpts.MempoolAuthAddresses }
		antehandler = ante.NewAnteHandler(app.accountKeeper, app.accountKeeper, auth.DefaultSigVerificationGasConsumer, getAuthorizedAddresses)
	} else {
		antehandler = ante.NewAnteHandler(app.accountKeeper, app.accountKeeper, auth.DefaultSigVerificationGasConsumer)
	}
	app.SetAnteHandler(antehandler)
	app.SetEndBlocker(app.EndBlocker)

	// load store
	if !appOpts.SkipLoadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// custom tx codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()

	ModuleBasics.RegisterCodec(cdc)
	vestingtypes.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)

	return cdc.Seal()
}

// SetBech32AddressPrefixes sets the global prefix to be used when serializing addresses to bech32 strings.
func SetBech32AddressPrefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32MainPrefix, Bech32MainPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator, Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
}

// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func SetBip44CoinType(config *sdk.Config) {
	config.SetCoinType(Bip44CoinType)
}

// application updates every end block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// custom logic for app initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	return app.mm.InitGenesis(ctx, genesisState)
}

// load a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range mAccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// TODO
// // BlacklistedAccAddrs returns all the app's module account addresses black listed for receiving tokens.
// func (app *App) BlacklistedAccAddrs() map[string]bool {
// 	blacklistedAddrs := make(map[string]bool)
// 	for acc := range mAccPerms {
// 		blacklistedAddrs[supply.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
// 	}

// 	return blacklistedAddrs
// }

// Codec returns the application's sealed codec.
func (app *App) Codec() *codec.Codec {
	return app.cdc
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// GetMaccPerms returns a mapping of the application's module account permissions.
func GetMaccPerms() map[string][]string {
	perms := make(map[string][]string)
	for k, v := range mAccPerms {
		perms[k] = v
	}
	return perms
}
