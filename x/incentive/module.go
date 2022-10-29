package incentive

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/incentive/client/cli"
	"github.com/kava-labs/kava/x/incentive/client/rest"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// ConsensusVersion defines the current module consensus version.
const ConsensusVersion = 2

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	// _ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the incentive module.
type AppModuleBasic struct{}

// Name returns the incentive module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec register module codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the incentive
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}

// ValidateGenesis performs genesis state validation for the incentive module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return err
	}
	return gs.Validate()
}

// RegisterInterfaces implements InterfaceModule.RegisterInterfaces
func (a AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterRESTRoutes registers REST routes for the incentive module.
func (a AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the incentive module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// TODO:
	// if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
	// 	panic(err)
	// }
}

// LegacyQuerierHandler returns sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}

// GetTxCmd returns the root tx command for the incentive module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns no root query command for the incentive module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements the sdk.AppModule interface.
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	cdpKeeper     types.CdpKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper, ck types.CdpKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
		cdpKeeper:      ck,
	}
}

// Name returns the incentive module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the incentive module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the incentive module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the incentive module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	// TODO: types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper, am.accountKeeper, am.bankKeeper))
}

// InitGenesis performs genesis initialization for the incentive module. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, am.accountKeeper, am.bankKeeper, am.cdpKeeper, genState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the incentive module
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(&gs)
}

// BeginBlock returns the begin blocker for the incentive module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock returns the end blocker for the incentive module. It returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//____________________________________________________________________________

// // RegisterStoreDecoder registers a decoder for incentive module's types
// func (AppModuleBasic) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
// sdr[types.StoreKey] = simulation.DecodeStore
// }

// // GenerateGenesisState creates a randomized GenState of the incentive module
// func (AppModuleBasic) GenerateGenesisState(simState *module.SimulationState) {
// simulation.RandomizedGenState(simState)
// }

// // RandomizedParams creates randomized incentive param changes for the simulator.
// func (AppModuleBasic) RandomizedParams(r *rand.Rand) []sim.ParamChange {
// return simulation.ParamChanges(r)
// }

// // ProposalContents doesn't return any content functions for governance proposals.
// func (AppModuleBasic) ProposalContents(_ module.SimulationState) []sim.WeightedProposalContent {
// return nil
// }

// // WeightedOperations returns the all the incentive module operations with their respective weights.
// func (am AppModule) WeightedOperations(simState module.SimulationState) []sim.WeightedOperation {
// return simulation.WeightedOperations(simState.AppParams, simState.Cdc, am.accountKeeper, am.supplyKeeper, am.keeper)
// }
