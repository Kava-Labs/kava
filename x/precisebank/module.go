package precisebank

import (
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kava-labs/kava/x/precisebank/keeper"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// ConsensusVersion defines the current module consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic app module basics object
type AppModuleBasic struct{}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

// Name get module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// Registers legacy amino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(gs)
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

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for precisebank module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
}

// GetTxCmd returns precisebank module's root tx command.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns precisebank module's root query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for precisebank module.
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	keeper keeper.Keeper,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

// Name returns precisebank module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
}

// RegisterInvariants registers precisebank module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper, am.bankKeeper)
}

// InitGenesis performs precisebank module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, am.accountKeeper, am.bankKeeper, &genState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns precisebank module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// BeginBlock executes all ABCI BeginBlock logic respective to precisebank module.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to precisebank module. It
// returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
