package greet

import (
	"context"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/kava-labs/kava/x/greet/client/rest"
	"github.com/kava-labs/kava/x/greet/client/cli"
	"github.com/kava-labs/kava/x/greet/keeper"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino){
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (a AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}


func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey)
}

// -------------------------------------------------------------------------------------- // 
type AppModule struct{
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

func (AppModule) RegisterInvariants(ir sdk.InvariantRegistry) { }


func (AppModule) Route() sdk.Route {
	return sdk.Route{}
}


func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

func (AppModule) ConsensusVersion() uint64 {
	return 1
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}


func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}


func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {	}

func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}


// -----------------------------------------------------------------------------------------------
var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper keeper.Keeper
}
 
func (m msgServer) CreateGreet(c context.Context, msg *types.MsgCreateGreet) (*types.MsgCreateGreetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	m.keeper.CreateGreet(ctx, types.MsgCreateGreet{Owner: msg.Owner, Message: msg.Message})

	return &types.MsgCreateGreetResponse{}, nil
}


func NewMsgServerImpl(keeper keeper.Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}