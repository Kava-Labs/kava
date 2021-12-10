package app

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
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
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
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
	transfer "github.com/cosmos/ibc-go/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/modules/core"
	ibcclient "github.com/cosmos/ibc-go/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/modules/core/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmlog "github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app/ante"
	kavaparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/auction"
	auctionkeeper "github.com/kava-labs/kava/x/auction/keeper"
	auctiontypes "github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/bep3"
	bep3keeper "github.com/kava-labs/kava/x/bep3/keeper"
	bep3types "github.com/kava-labs/kava/x/bep3/types"
	"github.com/kava-labs/kava/x/cdp"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/committee"
	committeeclient "github.com/kava-labs/kava/x/committee/client"
	committeekeeper "github.com/kava-labs/kava/x/committee/keeper"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	"github.com/kava-labs/kava/x/hard"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive"
	incentivekeeper "github.com/kava-labs/kava/x/incentive/keeper"
	incentivetypes "github.com/kava-labs/kava/x/incentive/types"
	issuance "github.com/kava-labs/kava/x/issuance"
	issuancekeeper "github.com/kava-labs/kava/x/issuance/keeper"
	issuancetypes "github.com/kava-labs/kava/x/issuance/types"
	"github.com/kava-labs/kava/x/kavadist"
	kavadistclient "github.com/kava-labs/kava/x/kavadist/client"
	kavadistkeeper "github.com/kava-labs/kava/x/kavadist/keeper"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
	pricefeed "github.com/kava-labs/kava/x/pricefeed"
	pricefeedkeeper "github.com/kava-labs/kava/x/pricefeed/keeper"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
	"github.com/kava-labs/kava/x/swap"
	swapkeeper "github.com/kava-labs/kava/x/swap/keeper"
	swaptypes "github.com/kava-labs/kava/x/swap/types"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

const (
	appName     = "kava"
	upgradeName = "v44"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics manages simple versions of full app modules.
	// It's used for things such as codec registration and genesis file verification.
	ModuleBasics = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler,
			distrclient.ProposalHandler,
			upgradeclient.ProposalHandler,
			upgradeclient.CancelProposalHandler,
			ibcclientclient.UpdateClientProposalHandler,
			ibcclientclient.UpgradeProposalHandler,
			kavadistclient.ProposalHandler,
			committeeclient.ProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		vesting.AppModuleBasic{},
		kavadist.AppModuleBasic{},
		auction.AppModuleBasic{},
		issuance.AppModuleBasic{},
		bep3.AppModuleBasic{},
		pricefeed.AppModuleBasic{},
		swap.AppModuleBasic{},
		cdp.AppModuleBasic{},
		hard.AppModuleBasic{},
		committee.AppModuleBasic{},
		incentive.AppModuleBasic{},
		validatorvesting.AppModuleBasic{},
	)

	// module account permissions
	// If these are changed, the permissions stored in accounts
	// must also be migrated during a chain upgrade.
	mAccPerms = map[string][]string{
		authtypes.FeeCollectorName:      nil,
		distrtypes.ModuleName:           nil,
		minttypes.ModuleName:            {authtypes.Minter},
		stakingtypes.BondedPoolName:     {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:  {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:             {authtypes.Burner},
		ibctransfertypes.ModuleName:     {authtypes.Minter, authtypes.Burner},
		kavadisttypes.KavaDistMacc:      {authtypes.Minter},
		auctiontypes.ModuleName:         nil,
		issuancetypes.ModuleAccountName: {authtypes.Minter, authtypes.Burner},
		bep3types.ModuleName:            {authtypes.Burner, authtypes.Minter},
		swaptypes.ModuleName:            nil,
		cdptypes.ModuleName:             {authtypes.Minter, authtypes.Burner},
		cdptypes.LiquidatorMacc:         {authtypes.Minter, authtypes.Burner},
		hardtypes.ModuleAccountName:     {authtypes.Minter},
	}
)

// Verify app interface at compile time
// var _ simapp.App = (*App)(nil) // TODO
var _ servertypes.Application = (*App)(nil)

// Options bundles several configuration params for an App.
// The zero value can be used as a sensible default.
type Options struct {
	SkipLoadLatest        bool
	SkipUpgradeHeights    map[int64]bool
	SkipGenesisInvariants bool
	InvariantCheckPeriod  uint
	MempoolEnableAuth     bool
	MempoolAuthAddresses  []sdk.AccAddress
}

// App is the Kava ABCI application.
type App struct {
	*baseapp.BaseApp

	// codec
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey

	// keepers from all the modules
	accountKeeper    authkeeper.AccountKeeper
	bankKeeper       bankkeeper.Keeper
	capabilityKeeper *capabilitykeeper.Keeper
	stakingKeeper    stakingkeeper.Keeper
	mintKeeper       mintkeeper.Keeper
	distrKeeper      distrkeeper.Keeper
	govKeeper        govkeeper.Keeper
	paramsKeeper     paramskeeper.Keeper
	crisisKeeper     crisiskeeper.Keeper
	slashingKeeper   slashingkeeper.Keeper
	ibcKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	upgradeKeeper    upgradekeeper.Keeper
	evidenceKeeper   evidencekeeper.Keeper
	transferKeeper   ibctransferkeeper.Keeper
	kavadistKeeper   kavadistkeeper.Keeper
	auctionKeeper    auctionkeeper.Keeper
	issuanceKeeper   issuancekeeper.Keeper
	bep3Keeper       bep3keeper.Keeper
	pricefeedKeeper  pricefeedkeeper.Keeper
	swapKeeper       swapkeeper.Keeper
	cdpKeeper        cdpkeeper.Keeper
	hardKeeper       hardkeeper.Keeper
	committeeKeeper  committeekeeper.Keeper
	incentiveKeeper  incentivekeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// configurator
	configurator module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		stdlog.Printf("Failed to get home dir %v", err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".kava")
}

// NewApp returns a reference to an initialized App.
func NewApp(
	logger tmlog.Logger,
	db dbm.DB,
	homePath string,
	traceStore io.Writer,
	encodingConfig kavaparams.EncodingConfig,
	options Options,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {

	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey,
		upgradetypes.StoreKey, evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey, kavadisttypes.StoreKey, auctiontypes.StoreKey,
		issuancetypes.StoreKey, bep3types.StoreKey, pricefeedtypes.StoreKey,
		swaptypes.StoreKey, cdptypes.StoreKey, hardtypes.StoreKey,
		committeetypes.StoreKey, incentivetypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	var app = &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// init params keeper and subspaces
	app.paramsKeeper = paramskeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)
	authSubspace := app.paramsKeeper.Subspace(authtypes.ModuleName)
	bankSubspace := app.paramsKeeper.Subspace(banktypes.ModuleName)
	stakingSubspace := app.paramsKeeper.Subspace(stakingtypes.ModuleName)
	mintSubspace := app.paramsKeeper.Subspace(minttypes.ModuleName)
	distrSubspace := app.paramsKeeper.Subspace(distrtypes.ModuleName)
	slashingSubspace := app.paramsKeeper.Subspace(slashingtypes.ModuleName)
	govSubspace := app.paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	crisisSubspace := app.paramsKeeper.Subspace(crisistypes.ModuleName)
	kavadistSubspace := app.paramsKeeper.Subspace(kavadisttypes.ModuleName)
	auctionSubspace := app.paramsKeeper.Subspace(auctiontypes.ModuleName)
	issuanceSubspace := app.paramsKeeper.Subspace(issuancetypes.ModuleName)
	bep3Subspace := app.paramsKeeper.Subspace(bep3types.ModuleName)
	pricefeedSubspace := app.paramsKeeper.Subspace(pricefeedtypes.ModuleName)
	swapSubspace := app.paramsKeeper.Subspace(swaptypes.ModuleName)
	cdpSubspace := app.paramsKeeper.Subspace(cdptypes.ModuleName)
	hardSubspace := app.paramsKeeper.Subspace(hardtypes.ModuleName)
	incentiveSubspace := app.paramsKeeper.Subspace(incentivetypes.ModuleName)
	ibcSubspace := app.paramsKeeper.Subspace(ibchost.ModuleName)
	ibctransferSubspace := app.paramsKeeper.Subspace(ibctransfertypes.ModuleName)

	bApp.SetParamStore(
		app.paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable()),
	)
	app.capabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
	scopedIBCKeeper := app.capabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.capabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.capabilityKeeper.Seal()

	// add keepers
	app.accountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authSubspace,
		authtypes.ProtoBaseAccount,
		mAccPerms,
	)
	app.bankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.accountKeeper,
		bankSubspace,
		app.loadBlockedMaccAddrs(),
	)
	app.stakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		stakingSubspace,
	)
	app.mintKeeper = mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		mintSubspace,
		&app.stakingKeeper,
		app.accountKeeper,
		app.bankKeeper,
		authtypes.FeeCollectorName,
	)
	app.distrKeeper = distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		distrSubspace,
		app.accountKeeper,
		app.bankKeeper,
		&app.stakingKeeper,
		authtypes.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)
	app.slashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		keys[slashingtypes.StoreKey],
		&app.stakingKeeper,
		slashingSubspace,
	)
	app.crisisKeeper = crisiskeeper.NewKeeper(
		crisisSubspace,
		options.InvariantCheckPeriod,
		app.bankKeeper,
		authtypes.FeeCollectorName,
	)
	app.upgradeKeeper = upgradekeeper.NewKeeper(
		options.SkipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
	)
	app.evidenceKeeper = *evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		&app.stakingKeeper,
		app.slashingKeeper,
	)

	app.ibcKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		ibcSubspace,
		app.stakingKeeper,
		app.upgradeKeeper,
		scopedIBCKeeper,
	)

	// TODO No evidence router is added so all submit evidence msgs will fail. Should there be a router added?
	govRouter := govtypes.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.ibcKeeper.ClientKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).AddRoute(kavadisttypes.RouterKey, kavadist.NewCommunityPoolMultiSpendProposalHandler(app.kavadistKeeper)).
		AddRoute(committeetypes.RouterKey, committee.NewProposalHandler(app.committeeKeeper))
	app.govKeeper = govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		govSubspace,
		app.accountKeeper,
		app.bankKeeper,
		&app.stakingKeeper,
		govRouter,
	)

	app.transferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		ibctransferSubspace,
		app.ibcKeeper.ChannelKeeper,
		&app.ibcKeeper.PortKeeper,
		app.accountKeeper,
		app.bankKeeper,
		scopedTransferKeeper,
	)
	transferModule := transfer.NewAppModule(app.transferKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferModule)
	app.ibcKeeper.SetRouter(ibcRouter)

	app.kavadistKeeper = kavadistkeeper.NewKeeper(
		appCodec,
		keys[kavadisttypes.StoreKey],
		kavadistSubspace,
		app.bankKeeper,
		app.accountKeeper,
		app.distrKeeper,
		app.ModuleAccountAddrs(),
	)
	app.auctionKeeper = auctionkeeper.NewKeeper(
		appCodec,
		keys[auctiontypes.StoreKey],
		auctionSubspace,
		app.bankKeeper,
		app.accountKeeper,
	)
	app.issuanceKeeper = issuancekeeper.NewKeeper(
		appCodec,
		keys[issuancetypes.StoreKey],
		issuanceSubspace,
		app.accountKeeper,
		app.bankKeeper,
	)
	app.bep3Keeper = bep3keeper.NewKeeper(
		appCodec,
		keys[bep3types.StoreKey],
		app.bankKeeper,
		app.accountKeeper,
		bep3Subspace,
		app.ModuleAccountAddrs(),
	)
	app.pricefeedKeeper = pricefeedkeeper.NewKeeper(
		appCodec,
		keys[pricefeedtypes.StoreKey],
		pricefeedSubspace,
	)
	swapKeeper := swapkeeper.NewKeeper(
		appCodec,
		keys[swaptypes.StoreKey],
		swapSubspace,
		app.accountKeeper,
		app.bankKeeper,
	)
	cdpKeeper := cdpkeeper.NewKeeper(
		appCodec,
		keys[cdptypes.StoreKey],
		cdpSubspace,
		app.pricefeedKeeper,
		app.auctionKeeper,
		app.bankKeeper,
		app.accountKeeper,
		mAccPerms,
	)
	hardKeeper := hardkeeper.NewKeeper(
		appCodec,
		keys[hardtypes.StoreKey],
		hardSubspace,
		app.accountKeeper,
		app.bankKeeper,
		app.pricefeedKeeper,
		app.auctionKeeper,
	)

	app.incentiveKeeper = incentivekeeper.NewKeeper(
		appCodec,
		keys[incentivetypes.StoreKey],
		incentiveSubspace,
		app.bankKeeper,
		&cdpKeeper,
		&hardKeeper,
		app.accountKeeper,
		app.stakingKeeper,
		&swapKeeper,
	)

	// create committee keeper with router
	committeeGovRouter := govtypes.NewRouter()
	committeeGovRouter.
		AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper))
	// Note: the committee proposal handler is not registered on the committee router. This means committees cannot create or update other committees.
	// Adding the committee proposal handler to the router is possible but awkward as the handler depends on the keeper which depends on the handler.
	app.committeeKeeper = committeekeeper.NewKeeper(
		appCodec,
		keys[committeetypes.StoreKey],
		committeeGovRouter,
		app.paramsKeeper,
		app.accountKeeper,
		app.bankKeeper,
	)

	// register the staking hooks
	// NOTE: These keepers are passed by reference above, so they will contain these hooks.
	app.stakingKeeper = *(app.stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks(), app.incentiveKeeper.Hooks())))

	app.swapKeeper = *swapKeeper.SetHooks(app.incentiveKeeper.Hooks())
	app.cdpKeeper = *cdpKeeper.SetHooks(cdptypes.NewMultiCDPHooks(app.incentiveKeeper.Hooks()))
	app.hardKeeper = *hardKeeper.SetHooks(hardtypes.NewMultiHARDHooks(app.incentiveKeeper.Hooks()))

	// create the module manager (Note: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.)
	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, app.accountKeeper, nil),
		bank.NewAppModule(appCodec, app.bankKeeper, app.accountKeeper),
		capability.NewAppModule(appCodec, *app.capabilityKeeper),
		staking.NewAppModule(appCodec, app.stakingKeeper, app.accountKeeper, app.bankKeeper),
		mint.NewAppModule(appCodec, app.mintKeeper, app.accountKeeper),
		distr.NewAppModule(appCodec, app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		gov.NewAppModule(appCodec, app.govKeeper, app.accountKeeper, app.bankKeeper),
		params.NewAppModule(app.paramsKeeper),
		crisis.NewAppModule(&app.crisisKeeper, options.SkipGenesisInvariants),
		slashing.NewAppModule(appCodec, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		ibc.NewAppModule(app.ibcKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		transferModule,
		vesting.NewAppModule(app.accountKeeper, app.bankKeeper),
		params.NewAppModule(app.paramsKeeper),
		kavadist.NewAppModule(app.kavadistKeeper, app.accountKeeper),
		auction.NewAppModule(app.auctionKeeper, app.accountKeeper, app.bankKeeper),
		issuance.NewAppModule(app.issuanceKeeper, app.accountKeeper, app.bankKeeper),
		bep3.NewAppModule(app.bep3Keeper, app.accountKeeper, app.bankKeeper),
		pricefeed.NewAppModule(app.pricefeedKeeper, app.accountKeeper),
		validatorvesting.NewAppModule(app.bankKeeper),
		swap.NewAppModule(app.swapKeeper, app.accountKeeper),
		cdp.NewAppModule(app.cdpKeeper, app.accountKeeper, app.pricefeedKeeper, app.bankKeeper),
		hard.NewAppModule(app.hardKeeper, app.accountKeeper, app.bankKeeper, app.pricefeedKeeper),
		committee.NewAppModule(app.committeeKeeper, app.accountKeeper),
		incentive.NewAppModule(app.incentiveKeeper, app.accountKeeper, app.bankKeeper, app.cdpKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// Auction.BeginBlocker will close out expired auctions and pay debt back to cdp.
	// So it should be run before cdp.BeginBlocker which cancels out debt with stable and starts more auctions.
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName, // TODO why new evidence and staking begin blockers?
		stakingtypes.ModuleName,
		kavadisttypes.ModuleName,
		auctiontypes.ModuleName,
		issuancetypes.ModuleName,
		bep3types.ModuleName,
		cdptypes.ModuleName,
		hardtypes.ModuleName,
		committeetypes.ModuleName,
		incentivetypes.ModuleName,
		ibchost.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		pricefeedtypes.ModuleName,
	)

	app.mm.SetOrderInitGenesis( // TODO why the different order?
		capabilitytypes.ModuleName,
		authtypes.ModuleName, // loads all accounts - should run before any module with a module account
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		ibchost.ModuleName,
		genutiltypes.ModuleName, // genutils must occur after staking so that pools are properly initialized with tokens from genesis accounts.
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		auctiontypes.ModuleName,
		kavadisttypes.ModuleName,
		auctiontypes.ModuleName,
		issuancetypes.ModuleName,
		bep3types.ModuleName,
		pricefeedtypes.ModuleName,
		swaptypes.ModuleName,
		cdptypes.ModuleName,
		hardtypes.ModuleName,
		incentivetypes.ModuleName,
		committeetypes.ModuleName,
		crisistypes.ModuleName, // runs the invariants at genesis - should run after other modules
	)

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: This is not required for apps that don't use the simulator for fuzz testing
	// transactions.
	// TODO
	// app.sm = module.NewSimulationManager(
	// 	auth.NewAppModule(app.accountKeeper),
	// 	bank.NewAppModule(app.bankKeeper, app.accountKeeper),
	// 	gov.NewAppModule(app.govKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper),
	// 	mint.NewAppModule(app.mintKeeper),
	// 	distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
	// 	staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.accountKeeper, app.bankKeeper),
	// 	slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
	// )
	// app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize the app
	var fetchers []ante.AddressFetcher // TODO add bep3 authorized addresses
	if options.MempoolEnableAuth {
		fetchers = append(fetchers,
			func(sdk.Context) []sdk.AccAddress { return options.MempoolAuthAddresses },
			app.bep3Keeper.GetAuthorizedAddresses,
			app.pricefeedKeeper.GetAuthorizedAddresses,
		)
	}
	antehandler, err := ante.NewAnteHandler(
		app.accountKeeper,
		app.bankKeeper,
		nil,
		app.ibcKeeper.ChannelKeeper,
		encodingConfig.TxConfig.SignModeHandler(),
		authante.DefaultSigVerificationGasConsumer,
		fetchers...,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create antehandler: %s", err))
	}

	app.SetAnteHandler(antehandler)
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// load store
	if !options.SkipLoadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Sprintf("failed to load latest version: %s", err))
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper

	return app
}

// BeginBlocker contains app specific logic for the BeginBlock abci call.
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker contains app specific logic for the EndBlock abci call.
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer contains app specific logic for the InitChain abci call.
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	// TODO: upgrade keeper version map?
	// app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads the app state for a particular height.
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range mAccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// InterfaceRegistry returns the app's InterfaceRegistry.
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// SimulationManager implements the SimulationApp interface.
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// Register legacy REST routes
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	RegisterLegacyTxRoutes(clientCtx, apiSvr.Router)

	// Register GRPC Gateway routes
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		// TODO where should old and new swagger docs be served? Should files be embedded in the binary?
		panic("TODO: register swagger in app")
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
// It registers transaction related endpoints on the app's grpc server.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
// It registers the standard tendermint grpc endpoints on the app's grpc server.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// loadBlockedMaccAddrs returns a map indicating the blocked status of each module account address
func (app *App) loadBlockedMaccAddrs() map[string]bool {
	modAccAddrs := app.ModuleAccountAddrs()
	kavadistMaccAddr := app.accountKeeper.GetModuleAddress(kavadisttypes.ModuleName)
	for addr := range modAccAddrs {
		// Set the kavadist module account address as unblocked
		if addr == kavadistMaccAddr.String() {
			modAccAddrs[addr] = false
		}
	}
	return modAccAddrs
}

// GetMaccPerms returns a mapping of the application's module account permissions.
func GetMaccPerms() map[string][]string {
	perms := make(map[string][]string)
	for k, v := range mAccPerms {
		perms[k] = v
	}
	return perms
}
