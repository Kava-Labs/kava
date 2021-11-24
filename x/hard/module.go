package hard

import (
	"context"
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

	"github.com/kava-labs/kava/x/hard/keeper"
	"github.com/kava-labs/kava/x/hard/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	// _ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic app module basics object
type AppModuleBasic struct{}

// Name get module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec register module codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the hard
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}

// ValidateGenesis performs genesis state validation for the hard module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err != nil {
		return err
	}
	return gs.Validate()
}

// RegisterInterfaces implements InterfaceModule.RegisterInterfaces
func (a AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// RegisterRESTRoutes registers REST routes for the swap module.
func (a AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	// TODO: rest.RegisterRoutes(clientCtx, rtr)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the gov module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// LegacyQuerierHandler returns sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// GetTxCmd returns the root tx command for the hard module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
	// TODO: return cli.GetTxCmd()
}

// GetQueryCmd returns no root query command for the hard module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
	// TODO: return cli.GetQueryCmd(types.StoreKey)
}

//____________________________________________________________________________

// AppModule app module type
type AppModule struct {
	AppModuleBasic

	keeper          keeper.Keeper
	accountKeeper   types.AccountKeeper
	bankKeeper      types.BankKeeper
	pricefeedKeeper types.PricefeedKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper, pricefeedKeeper types.PricefeedKeeper) AppModule {
	return AppModule{
		AppModuleBasic:  AppModuleBasic{},
		keeper:          keeper,
		accountKeeper:   accountKeeper,
		bankKeeper:      bankKeeper,
		pricefeedKeeper: pricefeedKeeper,
	}
}

// Name module name
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants register module invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the hard module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the hard module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	// TODO: grpc_query
	// types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// InitGenesis performs genesis initialization for the hard module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, am.accountKeeper, genState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the hard
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(&gs)
}

// BeginBlock module begin-block
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock module end-block
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//____________________________________________________________________________

// // GenerateGenesisState creates a randomized GenState of the hard module
// func (AppModuleBasic) GenerateGenesisState(simState *module.SimulationState) {
// 	simulation.RandomizedGenState(simState)
// }

// // ProposalContents doesn't return any content functions for governance proposals.
// func (AppModuleBasic) ProposalContents(_ module.SimulationState) []sim.WeightedProposalContent {
// 	return nil
// }

// // RandomizedParams returns nil because hard has no params.
// func (AppModuleBasic) RandomizedParams(r *rand.Rand) []sim.ParamChange {
// 	return simulation.ParamChanges(r)
// }

// // RegisterStoreDecoder registers a decoder for hard module's types
// func (AppModuleBasic) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
// 	sdr[types.StoreKey] = simulation.DecodeStore
// }

// // WeightedOperations returns the all the hard module operations with their respective weights.
// func (am AppModule) WeightedOperations(simState module.SimulationState) []sim.WeightedOperation {
// 	return nil
// }
