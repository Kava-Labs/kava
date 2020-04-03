package bep3

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	sim "github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/bep3/client/cli"
	"github.com/kava-labs/kava/x/bep3/client/rest"
	"github.com/kava-labs/kava/x/bep3/simulation"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModuleSimulation{}
)

// ----------------------------------------------
// 				  AppModuleBasic
// ----------------------------------------------

// AppModuleBasic defines the basic application module used by the bep3 module.
type AppModuleBasic struct{}

// Name returns the bep3 module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the bep3 module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the bep3
// module.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the bep3 module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var gs GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		return err
	}
	return gs.Validate()
}

// RegisterRESTRoutes registers the REST routes for the bep3 module.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr)
}

// GetTxCmd returns the root tx command for the bep3 module.
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

// GetQueryCmd returns no root query command for the bep3 module.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(StoreKey, cdc)
}

// ----------------------------------------------
// 				AppModuleSimulation
// ----------------------------------------------

// AppModuleSimulation defines the module simulation functions used by the auction module.
type AppModuleSimulation struct{}

// RegisterStoreDecoder registers a decoder for distribution module's types
func (AppModuleSimulation) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[StoreKey] = simulation.DecodeStore
}

// GenerateGenesisState creates a randomized GenState of the distribution module.
func (AppModuleSimulation) GenerateGenesisState(simState *module.SimulationState) {
	// TODO: rand seed should be specified in simuation config
	locRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	simState.Rand = locRand
	simulation.RandomizedGenState(simState)
}

// RandomizedParams creates randomized distribution param changes for the simulator.
func (AppModuleSimulation) RandomizedParams(r *rand.Rand) []sim.ParamChange {
	locRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	return simulation.ParamChanges(locRand)
}

// TODO:
// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModuleSimulation) WeightedOperations(simState module.SimulationState) []sim.WeightedOperation {
	return simulation.WeightedOperations(simState.AppParams, simState.Cdc,
		am.accountKeeper, am.keeper)
}

// ProposalContents returns all the distribution content functions used to
// simulate governance proposals.
// func (am AppModuleSimulation) ProposalContents(_ module.SimulationState) []sim.WeightedProposalContent {
// 	return simulation.ProposalContents(am.keeper)
// }

// ----------------------------------------------
// 					AppModule
// ----------------------------------------------

// AppModule implements the sdk.AppModule interface.
type AppModule struct {
	AppModuleBasic
	AppModuleSimulation

	keeper       Keeper
	supplyKeeper SupplyKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper, supplyKeeper SupplyKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		supplyKeeper:   supplyKeeper,
	}
}

// Name returns the bep3 module's name.
func (AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants registers the bep3 module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the bep3 module.
func (AppModule) Route() string {
	return RouterKey
}

// NewHandler returns an sdk.Handler for the bep3 module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute returns the bep3 module's querier route name.
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler returns the bep3 module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// InitGenesis performs genesis initialization for the bep3 module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, am.supplyKeeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the bep3
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the bep3 module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock returns the end blocker for the bep3 module. It returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
