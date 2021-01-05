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
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"

	"github.com/kava-labs/kava/app/ante"
	"github.com/kava-labs/kava/x/auction"
	"github.com/kava-labs/kava/x/bep3"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/issuance"
	"github.com/kava-labs/kava/x/kavadist"
	"github.com/kava-labs/kava/x/pricefeed"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
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
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		validatorvesting.AppModuleBasic{},
		bank.AppModuleBasic{},
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
		upgrade.AppModuleBasic{},
		supply.AppModuleBasic{},
		evidence.AppModuleBasic{},
		auction.AppModuleBasic{},
		cdp.AppModuleBasic{},
		pricefeed.AppModuleBasic{},
		committee.AppModuleBasic{},
		bep3.AppModuleBasic{},
		kavadist.AppModuleBasic{},
		incentive.AppModuleBasic{},
		issuance.AppModuleBasic{},
		hard.AppModuleBasic{},
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
		auction.ModuleName:          nil,
		cdp.ModuleName:              {supply.Minter, supply.Burner},
		cdp.LiquidatorMacc:          {supply.Minter, supply.Burner},
		cdp.SavingsRateMacc:         {supply.Minter},
		bep3.ModuleName:             {supply.Minter, supply.Burner},
		kavadist.ModuleName:         {supply.Minter},
		issuance.ModuleAccountName:  {supply.Minter, supply.Burner},
		hard.LPAccount:              {supply.Minter, supply.Burner},
		hard.DelegatorAccount:       {supply.Minter, supply.Burner},
		hard.ModuleAccountName:      {supply.Minter, supply.Burner},
		hard.LiquidatorAccount:      {supply.Minter, supply.Burner},
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		distr.ModuleName: true,
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
	accountKeeper   auth.AccountKeeper
	bankKeeper      bank.Keeper
	supplyKeeper    supply.Keeper
	stakingKeeper   staking.Keeper
	slashingKeeper  slashing.Keeper
	mintKeeper      mint.Keeper
	distrKeeper     distr.Keeper
	govKeeper       gov.Keeper
	crisisKeeper    crisis.Keeper
	upgradeKeeper   upgrade.Keeper
	paramsKeeper    params.Keeper
	evidenceKeeper  evidence.Keeper
	vvKeeper        validatorvesting.Keeper
	auctionKeeper   auction.Keeper
	cdpKeeper       cdp.Keeper
	pricefeedKeeper pricefeed.Keeper
	committeeKeeper committee.Keeper
	bep3Keeper      bep3.Keeper
	kavadistKeeper  kavadist.Keeper
	incentiveKeeper incentive.Keeper
	issuanceKeeper  issuance.Keeper
	hardKeeper      hard.Keeper

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
		bam.MainStoreKey, auth.StoreKey, staking.StoreKey,
		supply.StoreKey, mint.StoreKey, distr.StoreKey, slashing.StoreKey,
		gov.StoreKey, params.StoreKey, upgrade.StoreKey, evidence.StoreKey,
		validatorvesting.StoreKey, auction.StoreKey, cdp.StoreKey, pricefeed.StoreKey,
		bep3.StoreKey, kavadist.StoreKey, incentive.StoreKey, issuance.StoreKey, committee.StoreKey,
		hard.StoreKey,
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
	app.paramsKeeper = params.NewKeeper(app.cdc, keys[params.StoreKey], tkeys[params.TStoreKey])
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	mintSubspace := app.paramsKeeper.Subspace(mint.DefaultParamspace)
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace).WithKeyTable(gov.ParamKeyTable())
	evidenceSubspace := app.paramsKeeper.Subspace(evidence.DefaultParamspace)
	crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	auctionSubspace := app.paramsKeeper.Subspace(auction.DefaultParamspace)
	cdpSubspace := app.paramsKeeper.Subspace(cdp.DefaultParamspace)
	pricefeedSubspace := app.paramsKeeper.Subspace(pricefeed.DefaultParamspace)
	bep3Subspace := app.paramsKeeper.Subspace(bep3.DefaultParamspace)
	kavadistSubspace := app.paramsKeeper.Subspace(kavadist.DefaultParamspace)
	incentiveSubspace := app.paramsKeeper.Subspace(incentive.DefaultParamspace)
	issuanceSubspace := app.paramsKeeper.Subspace(issuance.DefaultParamspace)
	hardSubspace := app.paramsKeeper.Subspace(hard.DefaultParamspace)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		keys[auth.StoreKey],
		authSubspace,
		auth.ProtoBaseAccount,
	)
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		bankSubspace,
		app.BlacklistedAccAddrs(),
	)
	app.supplyKeeper = supply.NewKeeper(
		app.cdc,
		keys[supply.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		mAccPerms,
	)
	stakingKeeper := staking.NewKeeper(
		app.cdc,
		keys[staking.StoreKey],
		app.supplyKeeper,
		stakingSubspace,
	)
	app.mintKeeper = mint.NewKeeper(
		app.cdc,
		keys[mint.StoreKey],
		mintSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		auth.FeeCollectorName,
	)
	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		keys[distr.StoreKey],
		distrSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		auth.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)
	app.slashingKeeper = slashing.NewKeeper(
		app.cdc,
		keys[slashing.StoreKey],
		&stakingKeeper,
		slashingSubspace,
	)
	app.crisisKeeper = crisis.NewKeeper(
		crisisSubspace,
		app.invCheckPeriod,
		app.supplyKeeper,
		auth.FeeCollectorName,
	)
	app.upgradeKeeper = upgrade.NewKeeper(
		appOpts.SkipUpgradeHeights,
		keys[upgrade.StoreKey],
		app.cdc,
	)

	// create evidence keeper with router
	evidenceKeeper := evidence.NewKeeper(
		app.cdc,
		keys[evidence.StoreKey],
		evidenceSubspace,
		&app.stakingKeeper,
		app.slashingKeeper,
	)
	evidenceRouter := evidence.NewRouter()
	evidenceKeeper.SetRouter(evidenceRouter)
	app.evidenceKeeper = *evidenceKeeper

	// create committee keeper with router
	committeeGovRouter := gov.NewRouter()
	committeeGovRouter.
		AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
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
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(upgrade.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper)).
		AddRoute(committee.RouterKey, committee.NewProposalHandler(app.committeeKeeper))
	app.govKeeper = gov.NewKeeper(
		app.cdc,
		keys[gov.StoreKey],
		govSubspace,
		app.supplyKeeper,
		&stakingKeeper,
		govRouter,
	)

	app.vvKeeper = validatorvesting.NewKeeper(
		app.cdc,
		keys[validatorvesting.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.supplyKeeper,
		&stakingKeeper,
	)
	app.pricefeedKeeper = pricefeed.NewKeeper(
		app.cdc,
		keys[pricefeed.StoreKey],
		pricefeedSubspace,
	)
	app.auctionKeeper = auction.NewKeeper(
		app.cdc,
		keys[auction.StoreKey],
		app.supplyKeeper,
		auctionSubspace,
	)
	cdpKeeper := cdp.NewKeeper(
		app.cdc,
		keys[cdp.StoreKey],
		cdpSubspace,
		app.pricefeedKeeper,
		app.auctionKeeper,
		app.supplyKeeper,
		app.accountKeeper,
		mAccPerms,
	)
	app.bep3Keeper = bep3.NewKeeper(
		app.cdc,
		keys[bep3.StoreKey],
		app.supplyKeeper,
		app.accountKeeper,
		bep3Subspace,
		app.ModuleAccountAddrs(),
	)
	app.kavadistKeeper = kavadist.NewKeeper(
		app.cdc,
		keys[kavadist.StoreKey],
		kavadistSubspace,
		app.supplyKeeper,
	)
	app.incentiveKeeper = incentive.NewKeeper(
		app.cdc,
		keys[incentive.StoreKey],
		incentiveSubspace,
		app.supplyKeeper,
		&cdpKeeper,
		app.accountKeeper,
	)
	app.issuanceKeeper = issuance.NewKeeper(
		app.cdc,
		keys[issuance.StoreKey],
		issuanceSubspace,
		app.accountKeeper,
		app.supplyKeeper,
	)
	hardKeeper := hard.NewKeeper(
		app.cdc,
		keys[hard.StoreKey],
		hardSubspace,
		app.accountKeeper,
		app.supplyKeeper,
		&stakingKeeper,
		app.pricefeedKeeper,
		app.auctionKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	app.cdpKeeper = *cdpKeeper.SetHooks(cdp.NewMultiCDPHooks(app.incentiveKeeper.Hooks()))

	app.hardKeeper = *hardKeeper.SetHooks(hard.NewMultiHARDHooks(app.incentiveKeeper.Hooks()))

	// create the module manager (Note: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.)
	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		gov.NewAppModule(app.govKeeper, app.accountKeeper, app.supplyKeeper),
		mint.NewAppModule(app.mintKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
		distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.supplyKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.supplyKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		validatorvesting.NewAppModule(app.vvKeeper, app.accountKeeper),
		auction.NewAppModule(app.auctionKeeper, app.accountKeeper, app.supplyKeeper),
		cdp.NewAppModule(app.cdpKeeper, app.accountKeeper, app.pricefeedKeeper, app.supplyKeeper),
		pricefeed.NewAppModule(app.pricefeedKeeper, app.accountKeeper),
		bep3.NewAppModule(app.bep3Keeper, app.accountKeeper, app.supplyKeeper),
		kavadist.NewAppModule(app.kavadistKeeper, app.supplyKeeper),
		incentive.NewAppModule(app.incentiveKeeper, app.accountKeeper, app.supplyKeeper),
		committee.NewAppModule(app.committeeKeeper, app.accountKeeper),
		issuance.NewAppModule(app.issuanceKeeper, app.accountKeeper, app.supplyKeeper),
		hard.NewAppModule(app.hardKeeper, app.supplyKeeper, app.pricefeedKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// Auction.BeginBlocker will close out expired auctions and pay debt back to cdp.
	// So it should be run before cdp.BeginBlocker which cancels out debt with stable and starts more auctions.
	app.mm.SetOrderBeginBlockers(
		upgrade.ModuleName, mint.ModuleName, distr.ModuleName, slashing.ModuleName,
		validatorvesting.ModuleName, kavadist.ModuleName, auction.ModuleName, cdp.ModuleName,
		bep3.ModuleName, incentive.ModuleName, committee.ModuleName, issuance.ModuleName, hard.ModuleName,
	)

	app.mm.SetOrderEndBlockers(crisis.ModuleName, gov.ModuleName, staking.ModuleName, pricefeed.ModuleName)

	app.mm.SetOrderInitGenesis(
		auth.ModuleName, // loads all accounts - should run before any module with a module account
		validatorvesting.ModuleName, distr.ModuleName,
		staking.ModuleName, bank.ModuleName, slashing.ModuleName,
		gov.ModuleName, mint.ModuleName, evidence.ModuleName,
		pricefeed.ModuleName, cdp.ModuleName, auction.ModuleName,
		bep3.ModuleName, kavadist.ModuleName, incentive.ModuleName, committee.ModuleName, issuance.ModuleName, hard.ModuleName,
		supply.ModuleName,  // calculates the total supply from account - should run after modules that modify accounts in genesis
		crisis.ModuleName,  // runs the invariants at genesis - should run after other modules
		genutil.ModuleName, // genutils must occur after staking so that pools are properly initialized with tokens from genesis accounts.
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
		gov.NewAppModule(app.govKeeper, app.accountKeeper, app.supplyKeeper),
		mint.NewAppModule(app.mintKeeper),
		distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.supplyKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.supplyKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
		pricefeed.NewAppModule(app.pricefeedKeeper, app.accountKeeper),
		cdp.NewAppModule(app.cdpKeeper, app.accountKeeper, app.pricefeedKeeper, app.supplyKeeper),
		auction.NewAppModule(app.auctionKeeper, app.accountKeeper, app.supplyKeeper),
		bep3.NewAppModule(app.bep3Keeper, app.accountKeeper, app.supplyKeeper),
		kavadist.NewAppModule(app.kavadistKeeper, app.supplyKeeper),
		incentive.NewAppModule(app.incentiveKeeper, app.accountKeeper, app.supplyKeeper),
		committee.NewAppModule(app.committeeKeeper, app.accountKeeper),
		issuance.NewAppModule(app.issuanceKeeper, app.accountKeeper, app.supplyKeeper),
		hard.NewAppModule(app.hardKeeper, app.supplyKeeper, app.pricefeedKeeper),
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
		antehandler = ante.NewAnteHandler(app.accountKeeper, app.supplyKeeper, auth.DefaultSigVerificationGasConsumer, app.bep3Keeper.GetAuthorizedAddresses, app.pricefeedKeeper.GetAuthorizedAddresses, getAuthorizedAddresses)
	} else {
		antehandler = ante.NewAnteHandler(app.accountKeeper, app.supplyKeeper, auth.DefaultSigVerificationGasConsumer)
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
	vesting.RegisterCodec(cdc)
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
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlacklistedAccAddrs returns all the app's module account addresses black listed for receiving tokens.
func (app *App) BlacklistedAccAddrs() map[string]bool {
	blacklistedAddrs := make(map[string]bool)
	for acc := range mAccPerms {
		blacklistedAddrs[supply.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blacklistedAddrs
}

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
