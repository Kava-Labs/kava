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
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"

	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/kavadist"
	"github.com/kava-labs/kava/x/pricefeed"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"

	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc-transfer"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
)

const (
	appName          = "kava"
	Bech32MainPrefix = "kava"
	Bip44CoinType    = 459 // see https://github.com/satoshilabs/slips/blob/master/slip-0044.md
)

var (
	// default home directories for expected binaries
	DefaultCLIHome  = os.ExpandEnv("$HOME/.kvcli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.kvd")

	// ModuleBasics manages simple versions of full app modules. It's used for things such as codec registration and genesis file verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		validatorvesting.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler, distr.ProposalHandler, committee.ProposalHandler,
			upgradeclient.ProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		auction.AppModuleBasic{},
		cdp.AppModuleBasic{},
		pricefeed.AppModuleBasic{},
		committee.AppModuleBasic{},
		bep3.AppModuleBasic{},
		kavadist.AppModuleBasic{},
		incentive.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
	)

	// module account permissions
	mAccPerms = map[string][]string{
		auth.FeeCollectorName:       nil,
		distr.ModuleName:            nil,
		mint.ModuleName:             {auth.Minter},
		staking.BondedPoolName:      {auth.Burner, auth.Staking},
		staking.NotBondedPoolName:   {auth.Burner, auth.Staking},
		gov.ModuleName:              {auth.Burner},
		validatorvesting.ModuleName: {auth.Burner},
		auction.ModuleName:          nil,
		cdp.ModuleName:              {auth.Minter, auth.Burner},
		cdp.LiquidatorMacc:          {auth.Minter, auth.Burner},
		cdp.SavingsRateMacc:         {auth.Minter},
		bep3.ModuleName:             nil,
		kavadist.ModuleName:         {auth.Minter},
		transfer.ModuleName:         {auth.Minter, auth.Burner},
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		distr.ModuleName: true,
	}
)

// Verify app interface at compile time
var _ simapp.App = (*App)(nil)

// App represents an extended ABCI application
type App struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// subspaces
	subspaces map[string]params.Subspace

	// keepers from all the modules
	accountKeeper    auth.AccountKeeper
	bankKeeper       bank.Keeper
	capabilityKeeper *capability.Keeper
	stakingKeeper    staking.Keeper
	slashingKeeper   slashing.Keeper
	mintKeeper       mint.Keeper
	distrKeeper      distr.Keeper
	govKeeper        gov.Keeper
	crisisKeeper     crisis.Keeper
	upgradeKeeper    upgrade.Keeper
	paramsKeeper     params.Keeper
	ibcKeeper        *ibc.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	evidenceKeeper   evidence.Keeper
	transferKeeper   transfer.Keeper
	vvKeeper         validatorvesting.Keeper
	auctionKeeper    auction.Keeper
	cdpKeeper        cdp.Keeper
	pricefeedKeeper  pricefeed.Keeper
	committeeKeeper  committee.Keeper
	bep3Keeper       bep3.Keeper
	kavadistKeeper   kavadist.Keeper
	incentiveKeeper  incentive.Keeper

	// make scoped keepers public for test purposes
	scopedIBCKeeper      capability.ScopedKeeper
	scopedTransferKeeper capability.ScopedKeeper

	// the module manager
	mm *module.Manager

	// simulation manager
	// sm *module.SimulationManager
}

// NewApp returns a reference to an initialized App.
func NewApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	skipUpgradeHeights map[int64]bool, home string, invCheckPeriod uint,
	baseAppOptions ...func(*bam.BaseApp)) *App {

	appCodec, cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(
		auth.StoreKey, bank.StoreKey, staking.StoreKey,
		mint.StoreKey, distr.StoreKey, slashing.StoreKey,
		gov.StoreKey, validatorvesting.StoreKey, auction.StoreKey,
		cdp.StoreKey, pricefeed.StoreKey, bep3.StoreKey,
		kavadist.StoreKey, incentive.StoreKey, committee.StoreKey,
		params.StoreKey, ibc.StoreKey, upgrade.StoreKey,
		evidence.StoreKey, transfer.StoreKey, capability.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(params.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capability.MemStoreKey)

	var app = &App{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tkeys:          tkeys,
		memKeys:        memKeys,
		subspaces:      make(map[string]params.Subspace),
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(appCodec, keys[params.StoreKey], tkeys[params.TStoreKey])
	app.subspaces[auth.ModuleName] = app.paramsKeeper.Subspace(auth.DefaultParamspace)
	app.subspaces[bank.ModuleName] = app.paramsKeeper.Subspace(bank.DefaultParamspace)
	app.subspaces[staking.ModuleName] = app.paramsKeeper.Subspace(staking.DefaultParamspace)
	app.subspaces[mint.ModuleName] = app.paramsKeeper.Subspace(mint.DefaultParamspace)
	app.subspaces[distr.ModuleName] = app.paramsKeeper.Subspace(distr.DefaultParamspace)
	app.subspaces[slashing.ModuleName] = app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	app.subspaces[gov.ModuleName] = app.paramsKeeper.Subspace(gov.DefaultParamspace).WithKeyTable(gov.ParamKeyTable())
	app.subspaces[crisis.ModuleName] = app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	app.subspaces[auction.ModuleName] = app.paramsKeeper.Subspace(auction.DefaultParamspace)
	app.subspaces[cdp.ModuleName] = app.paramsKeeper.Subspace(cdp.DefaultParamspace)
	app.subspaces[pricefeed.ModuleName] = app.paramsKeeper.Subspace(pricefeed.DefaultParamspace)
	app.subspaces[bep3.ModuleName] = app.paramsKeeper.Subspace(bep3.DefaultParamspace)
	app.subspaces[kavadist.ModuleName] = app.paramsKeeper.Subspace(kavadist.DefaultParamspace)
	app.subspaces[incentive.ModuleName] = app.paramsKeeper.Subspace(incentive.DefaultParamspace)

	// set the BaseApp's parameter store
	bApp.SetParamStore(app.paramsKeeper.Subspace(bam.Paramspace).WithKeyTable(std.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.capabilityKeeper = capability.NewKeeper(appCodec, keys[capability.StoreKey], memKeys[capability.MemStoreKey])
	scopedIBCKeeper := app.capabilityKeeper.ScopeToModule(ibc.ModuleName)
	scopedTransferKeeper := app.capabilityKeeper.ScopeToModule(transfer.ModuleName)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(
		appCodec,
		keys[auth.StoreKey],
		app.subspaces[auth.ModuleName],
		auth.ProtoBaseAccount,
		mAccPerms,
	)
	app.bankKeeper = bank.NewBaseKeeper(
		appCodec, keys[bank.StoreKey], app.accountKeeper, app.subspaces[bank.ModuleName], app.BlacklistedAccAddrs(),
	)
	stakingKeeper := staking.NewKeeper(
		appCodec,
		keys[staking.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.subspaces[staking.ModuleName],
	)
	app.mintKeeper = mint.NewKeeper(
		appCodec,
		keys[mint.StoreKey],
		app.subspaces[mint.ModuleName],
		&stakingKeeper,
		app.accountKeeper,
		app.bankKeeper,
		auth.FeeCollectorName,
	)
	app.distrKeeper = distr.NewKeeper(
		appCodec,
		keys[distr.StoreKey],
		app.subspaces[distr.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
		&stakingKeeper,
		auth.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)
	app.slashingKeeper = slashing.NewKeeper(
		appCodec,
		keys[slashing.StoreKey],
		&stakingKeeper,
		app.subspaces[slashing.ModuleName],
	)
	app.crisisKeeper = crisis.NewKeeper(
		app.subspaces[crisis.ModuleName],
		invCheckPeriod,
		app.bankKeeper,
		auth.FeeCollectorName,
	)
	app.upgradeKeeper = upgrade.NewKeeper(skipUpgradeHeights, keys[upgrade.StoreKey], appCodec, home)

	// create committee keeper with router
	committeeGovRouter := gov.NewRouter()
	committeeGovRouter.
		AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(upgrade.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper))
	// Note: the committee proposal handler is not registered on the committee router. This means committees cannot create or update other committees.
	// Adding the committee proposal handler to the router is possible but awkward as the handler depends on the keeper which depends on the handler.
	app.committeeKeeper = committee.NewKeeper(
		app.cdc,
		keys[committee.StoreKey],
		committeeGovRouter,
		app.paramsKeeper,
	)

	// create gov keeper with router
	govRouter := gov.NewRouter()
	govRouter.
		AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(upgrade.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper)).
		AddRoute(committee.RouterKey, committee.NewProposalHandler(app.committeeKeeper))
	app.govKeeper = gov.NewKeeper(
		appCodec,
		keys[gov.StoreKey],
		app.subspaces[gov.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
		&stakingKeeper,
		govRouter,
	)

	app.vvKeeper = validatorvesting.NewKeeper(
		app.cdc,
		keys[validatorvesting.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.bankKeeper,
		&stakingKeeper,
	)
	app.pricefeedKeeper = pricefeed.NewKeeper(
		app.cdc,
		keys[pricefeed.StoreKey],
		app.subspaces[pricefeed.ModuleName],
	)
	app.auctionKeeper = auction.NewKeeper(
		app.cdc,
		keys[auction.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.subspaces[auction.ModuleName],
	)
	app.cdpKeeper = cdp.NewKeeper(
		app.cdc,
		keys[cdp.StoreKey],
		app.subspaces[cdp.ModuleName],
		app.pricefeedKeeper,
		app.auctionKeeper,
		app.bankKeeper,
		app.accountKeeper,
		mAccPerms,
	)
	app.bep3Keeper = bep3.NewKeeper(
		app.cdc,
		keys[bep3.StoreKey],
		app.bankKeeper,
		app.accountKeeper,
		app.subspaces[bep3.ModuleName],
		app.ModuleAccountAddrs(),
	)
	app.kavadistKeeper = kavadist.NewKeeper(
		app.cdc,
		keys[kavadist.StoreKey],
		app.subspaces[kavadist.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
	)
	app.incentiveKeeper = incentive.NewKeeper(
		app.cdc,
		keys[incentive.StoreKey],
		app.subspaces[incentive.ModuleName],
		app.bankKeeper,
		app.cdpKeeper,
		app.accountKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	// Create IBC Keeper
	// TODO: remove amino codec dependency once Tendermint version is upgraded with protobuf changes
	app.ibcKeeper = ibc.NewKeeper(
		app.cdc, appCodec, keys[ibc.StoreKey], app.stakingKeeper, scopedIBCKeeper,
	)

	// Create Transfer Keepers
	app.transferKeeper = transfer.NewKeeper(
		appCodec, keys[transfer.StoreKey],
		app.ibcKeeper.ChannelKeeper, &app.ibcKeeper.PortKeeper,
		app.accountKeeper, app.bankKeeper,
		scopedTransferKeeper,
	)
	transferModule := transfer.NewAppModule(app.transferKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := port.NewRouter()
	ibcRouter.AddRoute(transfer.ModuleName, transferModule)
	app.ibcKeeper.SetRouter(ibcRouter)

	// create evidence keeper with router
	evidenceKeeper := evidence.NewKeeper(
		appCodec,
		keys[evidence.StoreKey],
		&app.stakingKeeper,
		app.slashingKeeper,
	)
	evidenceRouter := evidence.NewRouter().
		AddRoute(ibcclient.RouterKey, ibcclient.HandlerClientMisbehaviour(app.ibcKeeper.ClientKeeper))
	evidenceKeeper.SetRouter(evidenceRouter)
	app.evidenceKeeper = *evidenceKeeper

	// create the module manager (Note: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.)
	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(appCodec, app.accountKeeper),
		bank.NewAppModule(appCodec, app.bankKeeper, app.accountKeeper),
		capability.NewAppModule(appCodec, *app.capabilityKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		gov.NewAppModule(appCodec, app.govKeeper, app.accountKeeper, app.bankKeeper),
		mint.NewAppModule(appCodec, app.mintKeeper, app.accountKeeper),
		slashing.NewAppModule(appCodec, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		distr.NewAppModule(appCodec, app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		staking.NewAppModule(appCodec, app.stakingKeeper, app.accountKeeper, app.bankKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		validatorvesting.NewAppModule(app.vvKeeper, app.accountKeeper),
		auction.NewAppModule(app.auctionKeeper, app.accountKeeper, app.bankKeeper),
		cdp.NewAppModule(app.cdpKeeper, app.accountKeeper, app.pricefeedKeeper, app.bankKeeper),
		pricefeed.NewAppModule(app.pricefeedKeeper, app.bankKeeper, app.accountKeeper),
		bep3.NewAppModule(app.bep3Keeper, app.accountKeeper, app.bankKeeper),
		kavadist.NewAppModule(app.kavadistKeeper, app.accountKeeper, app.bankKeeper),
		incentive.NewAppModule(app.incentiveKeeper, app.accountKeeper, app.bankKeeper),
		committee.NewAppModule(app.committeeKeeper, app.accountKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		ibc.NewAppModule(app.ibcKeeper),
		params.NewAppModule(app.paramsKeeper),
		transferModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// Auction.BeginBlocker will close out expired auctions and pay debt back to cdp.
	// So it should be run before cdp.BeginBlocker which cancels out debt with stable and starts more auctions.
	app.mm.SetOrderBeginBlockers(
		upgrade.ModuleName, mint.ModuleName, distr.ModuleName, slashing.ModuleName,
		validatorvesting.ModuleName, kavadist.ModuleName, auction.ModuleName, cdp.ModuleName,
		bep3.ModuleName, incentive.ModuleName, committee.ModuleName, ibc.ModuleName,
	)

	app.mm.SetOrderEndBlockers(crisis.ModuleName, gov.ModuleName, staking.ModuleName, pricefeed.ModuleName)

	// TODO: old SetOrderInitGenesis
	// app.mm.SetOrderInitGenesis(
	// 	auth.ModuleName, // loads all accounts - should run before any module with a module account
	// 	validatorvesting.ModuleName, distr.ModuleName,
	// 	staking.ModuleName, bank.ModuleName, slashing.ModuleName,
	// 	gov.ModuleName, mint.ModuleName, evidence.ModuleName,
	// 	pricefeed.ModuleName, cdp.ModuleName, auction.ModuleName,
	// 	bep3.ModuleName, kavadist.ModuleName, incentive.ModuleName,
	// 	committee.ModuleName, ibc.ModuleName, transfer.ModuleName,
	// 	bank.ModuleName,    // calculates the total bank from account - should run after modules that modify accounts in genesis
	// 	crisis.ModuleName,  // runs the invariants at genesis - should run after other modules
	// 	genutil.ModuleName, // genutils must occur after staking so that pools are properly initialized with tokens from genesis accounts.
	// )

	app.mm.SetOrderInitGenesis(
		capability.ModuleName, auth.ModuleName, validatorvesting.ModuleName, distr.ModuleName,
		staking.ModuleName, bank.ModuleName, slashing.ModuleName, gov.ModuleName, mint.ModuleName,
		crisis.ModuleName, ibc.ModuleName, genutil.ModuleName, pricefeed.ModuleName, cdp.ModuleName,
		auction.ModuleName, bep3.ModuleName, kavadist.ModuleName, incentive.ModuleName, committee.ModuleName,
		evidence.ModuleName, transfer.ModuleName,
	)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: This is not required for apps that don't use the simulator for fuzz testing
	// transactions.
	// app.sm = module.NewSimulationManager(
	// 	auth.NewAppModule(app.accountKeeper),
	// 	validatorvesting.NewAppModule(app.vvKeeper, app.accountKeeper),
	// 	bank.NewAppModule(app.bankKeeper, app.accountKeeper),
	// 	capability.NewAppModule(appCodec, *app.capabilityKeeper),
	// 	bank.NewAppModule(app.bankKeeper, app.accountKeeper),
	// 	gov.NewAppModule(app.govKeeper, app.accountKeeper, app.bankKeeper),
	// 	mint.NewAppModule(app.mintKeeper),
	// 	distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
	// 	staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.bankKeeper),
	// 	slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
	// 	pricefeed.NewAppModule(app.pricefeedKeeper, app.accountKeeper),
	// 	cdp.NewAppModule(app.cdpKeeper, app.accountKeeper, app.pricefeedKeeper, app.bankKeeper),
	// 	auction.NewAppModule(app.auctionKeeper, app.accountKeeper, app.bankKeeper),
	// 	bep3.NewAppModule(app.bep3Keeper, app.accountKeeper, app.bankKeeper),
	// 	kavadist.NewAppModule(app.kavadistKeeper, app.bankKeeper),
	// 	incentive.NewAppModule(app.incentiveKeeper, app.accountKeeper, app.bankKeeper),
	// 	committee.NewAppModule(app.committeeKeeper, app.accountKeeper),
	// 	evidence.NewAppModule(app.evidenceKeeper),
	// )

	// app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize the app
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(
		auth.NewAnteHandler(
			app.accountKeeper, app.bankKeeper, *app.ibcKeeper,
			auth.DefaultSigVerificationGasConsumer,
		),
	)
	app.SetEndBlocker(app.EndBlocker)

	// load store
	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	ctx := app.BaseApp.NewUncachedContext(true, abci.Header{})
	app.capabilityKeeper.InitializeAndSeal(ctx)

	app.scopedIBCKeeper = scopedIBCKeeper
	app.scopedTransferKeeper = scopedTransferKeeper

	return app
}

// custom tx codec
func MakeCodec() (*std.Codec, *codec.Codec) {

	cdc := std.MakeCodec(ModuleBasics)
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	appCodec := std.NewAppCodec(cdc, interfaceRegistry)

	sdk.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaceModules(interfaceRegistry)

	return appCodec, cdc
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

// TODO:
// // custom logic for app initialization
// func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
// 	var genesisState GenesisState
// 	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
// 	return app.mm.InitGenesis(ctx, app.cdc, genesisState)
// }

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState simapp.GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	return app.mm.InitGenesis(ctx, app.cdc, genesisState)
}

// load a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range mAccPerms {
		modAccAddrs[auth.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlacklistedAccAddrs returns all the app's module account addresses black listed for receiving tokens.
func (app *App) BlacklistedAccAddrs() map[string]bool {
	blacklistedAddrs := make(map[string]bool)
	for acc := range mAccPerms {
		blacklistedAddrs[auth.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blacklistedAddrs
}

// Codec returns the application's sealed codec.
func (app *App) Codec() *codec.Codec {
	return app.cdc
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	// return app.sm
	return nil
}

// GetMaccPerms returns a mapping of the application's module account permissions.
func GetMaccPerms() map[string][]string {
	perms := make(map[string][]string)
	for k, v := range mAccPerms {
		perms[k] = v
	}
	return perms
}
