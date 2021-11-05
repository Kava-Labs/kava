package cdp

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

	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
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

// DefaultGenesis returns default genesis state as raw bytes for the cdp
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}

// ValidateGenesis module validate genesis
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
	// TODO:
	// rest.RegisterRoutes(clientCtx, rtr)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the gov module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// LegacyQuerierHandler returns no sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// GetTxCmd returns the root tx command for the cdp module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
	// TODO:
	// return cli.GetTxCmd()
}

// GetQueryCmd returns the root query command for the auction module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
	// TODO: return cli.GetQueryCmd(StoreKey)
}

//____________________________________________________________________________

// AppModule app module type
type AppModule struct {
	AppModuleBasic

	keeper          keeper.Keeper
	accountKeeper   types.AccountKeeper
	pricefeedKeeper types.PricefeedKeeper
	bankKeeper      types.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper, accountKeeper types.AccountKeeper, pricefeedKeeper types.PricefeedKeeper, bankKeeper types.BankKeeper) AppModule {
	return AppModule{
		AppModuleBasic:  AppModuleBasic{},
		keeper:          keeper,
		accountKeeper:   accountKeeper,
		pricefeedKeeper: pricefeedKeeper,
		bankKeeper:      bankKeeper,
	}
}

// Name module name
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants register module invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route module message route name
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// NewHandler module handler
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute module querier route name
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(am.keeper))
}

// InitGenesis module init-genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, am.pricefeedKeeper, am.accountKeeper, genState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(&gs)
}

// BeginBlock module begin-block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

// EndBlock module end-block
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

//____________________________________________________________________________

// // GenerateGenesisState creates a randomized GenState of the cdp module
// func (AppModuleBasic) GenerateGenesisState(simState *module.SimulationState) {
// 	simulation.RandomizedGenState(simState)
// }

// // ProposalContents doesn't return any content functions for governance proposals.
// func (AppModuleBasic) ProposalContents(_ module.SimulationState) []sim.WeightedProposalContent {
// 	return nil
// }

// // RandomizedParams returns nil because cdp has no params.
// func (AppModuleBasic) RandomizedParams(r *rand.Rand) []sim.ParamChange {
// 	return simulation.ParamChanges(r)
// }

// // RegisterStoreDecoder registers a decoder for cdp module's types
// func (AppModuleBasic) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
// 	sdr[StoreKey] = simulation.DecodeStore
// }

// // WeightedOperations returns the all the cdp module operations with their respective weights.
// func (am AppModule) WeightedOperations(simState module.SimulationState) []sim.WeightedOperation {
// 	return simulation.WeightedOperations(simState.AppParams, simState.Cdc, am.accountKeeper, am.keeper, am.pricefeedKeeper)
// }
