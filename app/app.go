package app

import (
	"io"
	"os"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/liquidator"
	"github.com/kava-labs/kava/x/pricefeed"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

const (
	appName          = "kava"
	Bech32MainPrefix = "kava"
)

var (
	// default home directories for expected binaries
	DefaultCLIHome  = os.ExpandEnv("$HOME/.kvcli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.kvd")

	// ModuleBasics manages simple versions of full app modules. It's used for things such as codec registration and genesis file verification.
	ModuleBasics = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		validatorvesting.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(paramsclient.ProposalHandler, distr.ProposalHandler),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},
		auction.AppModuleBasic{},
		cdp.AppModuleBasic{},
		liquidator.AppModuleBasic{},
		pricefeed.AppModuleBasic{},
	)

	// module account permissions
	mAccPerms = map[string][]string{
		auth.FeeCollectorName:       nil,
		distr.ModuleName:            nil,
		mint.ModuleName:             {supply.Minter},
		staking.BondedPoolName:      {supply.Burner, supply.Staking},
		staking.NotBondedPoolName:   {supply.Burner, supply.Staking},
		gov.ModuleName:              {supply.Burner},
		validatorvesting.ModuleName: {supply.Burner},
	}
)

// Extended ABCI application
type App struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	// keepers from all the modules
	accountKeeper    auth.AccountKeeper
	bankKeeper       bank.Keeper
	supplyKeeper     supply.Keeper
	stakingKeeper    staking.Keeper
	slashingKeeper   slashing.Keeper
	mintKeeper       mint.Keeper
	distrKeeper      distr.Keeper
	govKeeper        gov.Keeper
	crisisKeeper     crisis.Keeper
	paramsKeeper     params.Keeper
	vvKeeper         validatorvesting.Keeper
	auctionKeeper    auction.Keeper
	cdpKeeper        cdp.Keeper
	liquidatorKeeper liquidator.Keeper
	pricefeedKeeper  pricefeed.Keeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager
}

// NewApp returns a reference to an initialized App.
func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint,
	baseAppOptions ...func(*bam.BaseApp)) *App {

	cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(
		bam.MainStoreKey, auth.StoreKey, staking.StoreKey,
		supply.StoreKey, mint.StoreKey, distr.StoreKey, slashing.StoreKey,
		gov.StoreKey, params.StoreKey, validatorvesting.StoreKey,
		auction.StoreKey, cdp.StoreKey, liquidator.StoreKey, pricefeed.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(params.TStoreKey)

	var app = &App{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tkeys:          tkeys,
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(app.cdc, keys[params.StoreKey], tkeys[params.TStoreKey], params.DefaultCodespace)
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := app.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace).WithKeyTable(gov.ParamKeyTable())
	crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	auctionSubspace := app.paramsKeeper.Subspace(auction.DefaultParamspace)
	cdpSubspace := app.paramsKeeper.Subspace(cdp.DefaultParamspace)
	//liquidatorSubspace := app.paramsKeeper.Subspace(liquidator.DefaultParamspace)
	pricefeedSubspace := app.paramsKeeper.Subspace(pricefeed.DefaultParamspace)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		keys[auth.StoreKey],
		authSubspace,
		auth.ProtoBaseAccount)
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		bankSubspace,
		bank.DefaultCodespace,
		app.ModuleAccountAddrs())
	app.supplyKeeper = supply.NewKeeper(
		app.cdc,
		keys[supply.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		mAccPerms)
	stakingKeeper := staking.NewKeeper(
		app.cdc,
		keys[staking.StoreKey],
		app.supplyKeeper,
		stakingSubspace,
		staking.DefaultCodespace)
	app.mintKeeper = mint.NewKeeper(
		app.cdc,
		keys[mint.StoreKey],
		mintSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		auth.FeeCollectorName)
	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		keys[distr.StoreKey],
		distrSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		distr.DefaultCodespace,
		auth.FeeCollectorName,
		app.ModuleAccountAddrs())
	app.slashingKeeper = slashing.NewKeeper(
		app.cdc,
		keys[slashing.StoreKey],
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
		keys[gov.StoreKey],
		govSubspace,
		app.supplyKeeper,
		&stakingKeeper,
		gov.DefaultCodespace,
		govRouter)
	app.vvKeeper = validatorvesting.NewKeeper(
		app.cdc,
		keys[validatorvesting.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.supplyKeeper,
		&stakingKeeper)
	app.pricefeedKeeper = pricefeed.NewKeeper(
		app.cdc,
		keys[pricefeed.StoreKey],
		pricefeedSubspace,
		pricefeed.DefaultCodespace)
	app.cdpKeeper = cdp.NewKeeper(
		app.cdc,
		keys[cdp.StoreKey],
		cdpSubspace,
		app.pricefeedKeeper,
		app.bankKeeper)
	app.auctionKeeper = auction.NewKeeper(
		app.cdc,
		keys[auction.StoreKey],
		app.supplyKeeper, // CDP keeper standing in for bank
		auctionSubspace)
	// app.liquidatorKeeper = liquidator.NewKeeper(
	// 	app.cdc,
	// 	keys[liquidator.StoreKey],
	// 	liquidatorSubspace,
	// 	app.cdpKeeper,
	// 	app.auctionKeeper,
	// 	app.cdpKeeper) // CDP keeper standing in for bank

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	// create the module manager (Note: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.)
	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		distr.NewAppModule(app.distrKeeper, app.supplyKeeper),
		gov.NewAppModule(app.govKeeper, app.supplyKeeper),
		mint.NewAppModule(app.mintKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.supplyKeeper),
		validatorvesting.NewAppModule(app.vvKeeper, app.accountKeeper),
		auction.NewAppModule(app.auctionKeeper),
		cdp.NewAppModule(app.cdpKeeper, app.pricefeedKeeper),
		liquidator.NewAppModule(app.liquidatorKeeper),
		pricefeed.NewAppModule(app.pricefeedKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(mint.ModuleName, distr.ModuleName, slashing.ModuleName, validatorvesting.ModuleName)

	app.mm.SetOrderEndBlockers(crisis.ModuleName, gov.ModuleName, staking.ModuleName, pricefeed.ModuleName, auction.ModuleName) // TODO is this correct order?

	// Note: genutils must occur after staking so that pools are properly
	// initialized with tokens from genesis accounts.
	//
	// Note: Changing the order of the auth module and modules that use module accounts
	// results in subtle changes to the way accounts are loaded from genesis.
	app.mm.SetOrderInitGenesis(
		auth.ModuleName, validatorvesting.ModuleName, distr.ModuleName,
		staking.ModuleName, bank.ModuleName, slashing.ModuleName,
		gov.ModuleName, mint.ModuleName, supply.ModuleName, crisis.ModuleName, genutil.ModuleName,
		pricefeed.ModuleName, cdp.ModuleName, auction.ModuleName, liquidator.ModuleName, // TODO is this order ok?
	)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: This is not required for apps that don't use the simulator for fuzz testing
	// transactions.
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(app.accountKeeper),
		validatorvesting.NewAppModule(app.vvKeeper, app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		gov.NewAppModule(app.govKeeper, app.supplyKeeper),
		mint.NewAppModule(app.mintKeeper),
		distr.NewAppModule(app.distrKeeper, app.supplyKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.supplyKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	// initialize the app
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.supplyKeeper, auth.DefaultSigVerificationGasConsumer))
	app.SetEndBlocker(app.EndBlocker)

	// load store
	if loadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
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
	vesting.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)

	return cdc.Seal()
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
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range mAccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// Codec returns the application's sealed codec.
func (app *App) Codec() *codec.Codec {
	return app.cdc
}

// GetMaccPerms returns a mapping of the application's module account permissions.
func GetMaccPerms() map[string][]string {
	perms := make(map[string][]string)
	for k, v := range mAccPerms {
		perms[k] = v
	}
	return perms
}
