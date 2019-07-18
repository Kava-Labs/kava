package app

import (
	"io"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// TODO investigate monkey patch import

const (
	appName          = "kava"
	Bech32MainPrefix = "kava"
)

var (
	// default home directories for expected binaries
	DefaultCLIHome  = os.ExpandEnv("$HOME/.kvcli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.kvd")

	// _ manages simple versions of full app modules. It's used for things such as codec registration and genesis file verification
	ModuleBasics module.BasicManager // TODO rename

	// module account permissions
	mAccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		distr.ModuleName:          nil,
		mint.ModuleName:           {supply.Minter},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		gov.ModuleName:            {supply.Burner},
	}
)

func init() {
	ModuleBasics = module.NewBasicManager(
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(paramsclient.ProposalHandler, distrclient.ProposalHandler),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},
	)
}

// Extended ABCI application
type App struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keyMain     *sdk.KVStoreKey
	keyAccount  *sdk.KVStoreKey
	keySupply   *sdk.KVStoreKey
	keyStaking  *sdk.KVStoreKey
	tkeyStaking *sdk.TransientStoreKey
	keySlashing *sdk.KVStoreKey
	keyMint     *sdk.KVStoreKey
	keyDistr    *sdk.KVStoreKey
	tkeyDistr   *sdk.TransientStoreKey
	keyGov      *sdk.KVStoreKey
	keyParams   *sdk.KVStoreKey
	tkeyParams  *sdk.TransientStoreKey

	// Manage getting and setting accounts
	accountKeeper  auth.AccountKeeper
	bankKeeper     bank.Keeper
	supplyKeeper   supply.Keeper
	stakingKeeper  staking.Keeper
	slashingKeeper slashing.Keeper
	mintKeeper     mint.Keeper
	distrKeeper    distr.Keeper
	govKeeper      gov.Keeper
	crisisKeeper   crisis.Keeper
	paramsKeeper   params.Keeper

	// the module manager
	mm *module.Manager
}

// NewApp returns a reference to an initialized App.
func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint,
	baseAppOptions ...func(*bam.BaseApp)) *App {

	cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	var app = &App{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keyMain:        sdk.NewKVStoreKey(bam.MainStoreKey),
		keyAccount:     sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:      sdk.NewKVStoreKey(supply.StoreKey),
		keyStaking:     sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking:    sdk.NewTransientStoreKey(staking.TStoreKey),
		keyMint:        sdk.NewKVStoreKey(mint.StoreKey),
		keyDistr:       sdk.NewKVStoreKey(distr.StoreKey),
		tkeyDistr:      sdk.NewTransientStoreKey(distr.TStoreKey),
		keySlashing:    sdk.NewKVStoreKey(slashing.StoreKey),
		keyGov:         sdk.NewKVStoreKey(gov.StoreKey),
		keyParams:      sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:     sdk.NewTransientStoreKey(params.TStoreKey),
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams, params.DefaultCodespace)
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := app.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace)
	crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		authSubspace,
		auth.ProtoBaseAccount)
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		bankSubspace,
		bank.DefaultCodespace)
	app.supplyKeeper = supply.NewKeeper(
		app.cdc,
		app.keySupply,
		app.accountKeeper,
		app.bankKeeper,
		supply.DefaultCodespace,
		mAccPerms)
	stakingKeeper := staking.NewKeeper(
		app.cdc,
		app.keyStaking,
		app.tkeyStaking,
		app.supplyKeeper,
		stakingSubspace,
		staking.DefaultCodespace)
	app.mintKeeper = mint.NewKeeper(
		app.cdc,
		app.keyMint,
		mintSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		auth.FeeCollectorName)
	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		app.keyDistr,
		distrSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		distr.DefaultCodespace,
		auth.FeeCollectorName)
	app.slashingKeeper = slashing.NewKeeper(
		app.cdc,
		app.keySlashing,
		&stakingKeeper,
		slashingSubspace,
		slashing.DefaultCodespace)
	app.crisisKeeper = crisis.NewKeeper(
		crisisSubspace,
		invCheckPeriod,
		app.supplyKeeper,
		auth.FeeCollectorName)
	govRouter := gov.NewRouter()
	govRouter.
		AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper))
	app.govKeeper = gov.NewKeeper(
		app.cdc,
		app.keyGov,
		app.paramsKeeper,
		govSubspace,
		app.supplyKeeper,
		&stakingKeeper,
		gov.DefaultCodespace,
		govRouter)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	// create the module manager
	app.mm = module.NewManager(
		genaccounts.NewAppModule(app.accountKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(app.crisisKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		distr.NewAppModule(app.distrKeeper, app.supplyKeeper),
		gov.NewAppModule(app.govKeeper, app.supplyKeeper),
		mint.NewAppModule(app.mintKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.distrKeeper, app.accountKeeper, app.supplyKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(mint.ModuleName, distr.ModuleName, slashing.ModuleName)

	app.mm.SetOrderEndBlockers(gov.ModuleName, staking.ModuleName)

	// genutils must occur after staking so that pools are properly
	// initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(genaccounts.ModuleName, distr.ModuleName,
		staking.ModuleName, auth.ModuleName, bank.ModuleName, slashing.ModuleName,
		gov.ModuleName, mint.ModuleName, supply.ModuleName, crisis.ModuleName, genutil.ModuleName)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// initialize the app
	app.MountStores(app.keyMain, app.keyAccount, app.keySupply, app.keyStaking,
		app.keyMint, app.keyDistr, app.keySlashing, app.keyGov, app.keyParams,
		app.tkeyParams, app.tkeyStaking, app.tkeyDistr)
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.supplyKeeper, auth.DefaultSigVerificationGasConsumer))
	app.SetEndBlocker(app.EndBlocker)

	// load store
	if loadLatest {
		err := app.LoadLatestVersion(app.keyMain)
		if err != nil {
			cmn.Exit(err.Error())
		}
	}

	return app
}

// custom tx codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

func SetBech32AddressPrefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32MainPrefix, Bech32MainPrefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator, Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, Bech32MainPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
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
	return app.LoadVersion(height, app.keyMain)
}

// TODO only used in sim test, needed?
// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range mAccPerms {
		modAccAddrs[app.supplyKeeper.GetModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
