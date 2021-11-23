#  Building a Module 

In this tutorial we will be going over building a module in Kava to show how easy it is to build on top of the Kava ecosystem. This module will be simple in nature but will show how to set up and connect a module to Kava and can be used as a starting point for more complex modules. 

Draft!:

 1. finish up explanation of each section 
 2. possibly split up sections into separate pages 
 3. add the set up and start up (a bit complicated for someone to follow along because master is not yet v44, and this tutorial is based on v44)

	
## Set up
```
a bit complicated for now becuase they have to clone a specific branch to work on v44
we can either have them do that or wait unit we are running v44 in master
```
## Defining protobuf Types 

The first step in building a new Kava Module is to define our Module's types. To do that we use Protocol Buffers which is a used for serializing structured data and generating code for multiple target languages, Protocol Buffers are also smaller than JSON & XML so sending data around the network will be less expensive. [Learn More](https://developers.google.com/protocol-buffers). 

Our Protobuf files will all live in ```proto/kava``` directory.  we will create a new directory with the new module ```greet``` and add the following files in the ```proto/greet/v1beta1/``` directory
```
├── ...
├── proto                 # Contains .proto files for all modules
│   ├── kava ├── greet ├── v1beta1 ├── |genesis.proto
|	...								   |greet.proto
|	...								   |query.proto
|	...								   |tx.proto
|   ... 
```
### Defining The Greet Type
Inside the ```proto/greet/v1beta1/greet.proto``` file lets define our greet type: 
```
syntax = "proto3";
package  kava.greet.v1beta1;
import  "cosmos_proto/cosmos.proto";
import  "gogoproto/gogo.proto";
option  go_package = "github.com/kava-labs/kava/x/greet/types";

message Greet {
string owner = 1;
string id = 2;
string message = 3;
}
```
Here we are saying that we have a Greet type that will have an owner, an id and a message that will contain the greet string. Once we have that define we are ready to set up a way to create this greet message and query it.  

### Creating a new Greeting 
Inside the ```proto/greet/v1beta1/tx.proto``` file lets define our Msg Type: 
```
syntax = "proto3";
package  kava.greet.v1beta1;
import  "gogoproto/gogo.proto";
import  "cosmos_proto/cosmos.proto";
option  go_package = "github.com/kava-labs/kava/x/greet/types";

// here we define our Msg Service which will handle our CreateGreet transaction
service  Msg {
// CreateGreet will take a MsgCreateGreet and return an MsgCreateGreetResponse response
rpc  CreateGreet (MsgCreateGreet) returns (MsgCreateGreetResponse);
}

// to create a greet message provide the message and the owner of the message *note Id will be auto-generated
message  MsgCreateGreet {
string message = 1;
string owner = 2;
}

// we will leave our response type empty 
message  MsgCreateGreetResponse { }
```
Now that we have defined how to create a new Greeting lets finish up by setting up our queries to view a specific greeting or all of them.

### Querying Greetings 
```
syntax = "proto3";
package  kava.greet.v1beta1;
option  go_package = "github.com/kava-labs/kava/x/greet/types";
import  "gogoproto/gogo.proto";
import  "google/api/annotations.proto";
import  "cosmos/base/query/v1beta1/pagination.proto";
import  "cosmos_proto/cosmos.proto";
import  "kava/greet/v1beta1/greet.proto";

// this is where we link up our query services

service  Query {
	// Greet will take and input of type QueryGetGreetRequest and return QueryGetGreetResponse

	rpc  Greet(QueryGetGreetRequest) returns (QueryGetGreetResponse) {
	// this is the endpoint for our Greet service
	option  (google.api.http).get = "/kava/greet/v1beta1/greetings/{id}";
	}

	// GreetAll will show all of the greet messages we have added and will take a QueryAllGreetRequest and return a QueryAllGreetResponse
	rpc  GreetAll(QueryAllGreetRequest) returns (QueryAllGreetResponse) {
	// this is the endpoint to see all of the greetings in our chain
	option  (google.api.http).get = "/kava/swap/v1beta1/greetings";
	}
}

// the Greet Request will need an Id of the greet message
message  QueryGetGreetRequest {
string id = 1;
}

// the Greet Response will return the Greet Type which we defined in the greet.proto
message  QueryGetGreetResponse {
Greet greeting = 1;
}

// the query all request will not require any fields except for the pagination incase we have a lot 
message  QueryAllGreetRequest {
cosmos.base.query.v1beta1.PageRequest pagination = 1;
}

// will return a list of Greet types
message  QueryAllGreetResponse {
repeated  Greet greetings = 1;
cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```

Once we have now defined our query, tx, and greet proto files we finally need to set up the genesis.proto file and then we are ready to generate these types in the genesis file we will create a basic genesis that doesn't do anything for now. 
```
syntax = "proto3";
package  kava.greet.v1beta1;
import  "kava/greet/v1beta1/greet.proto";
import  "gogoproto/gogo.proto";
import  "google/protobuf/timestamp.proto";
import  "cosmos_proto/cosmos.proto";
option  go_package = "github.com/kava-labs/kava/x/greet/types";
// our gensis state message will be empty for this tutorial
message  GenesisState {}
```
Once all the files are filled in we are ready to generate our proto types. in the Kava Directory run ```make proto-gen ``` to generate the types, this will create a folder inside the ```x/greet``` and will contain the auto-generated proto types. 

### Developing Our Greet Module

First thing we will do is set up our greet module's keys we will create a new file ```x/greet/types/keys.go```, and populate this with keys for the greet module, we will use these keys later as we build our module. 
```
package  types

const (
	ModuleName =  "greet"
	StoreKey = ModuleName
	RouterKey = ModuleName
	QuerierRoute = ModuleName
	GreetKey =  "Greet-value-"
	GreetCountKey =  "Greet-count-"
)

func  KeyPrefix(p string) []byte {
	return []byte(p)
}
```

Once we have our keys set up let's create our codec which will handle serialization of our data and register our protobuf types and interfaces.

Create   ```x/greet/types/codec.go``` file and add the code below: 
```
package  types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdk  "github.com/cosmos/cosmos-sdk/types"
)

func  RegisterCodec(cdc *codec.LegacyAmino){
	cdc.RegisterConcrete(&MsgCreateGreet{}, "greet/CreateGreet", nil)
}

func  RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateGreet{}, "greet/CreateGreet", nil)
}

func  RegisterInterfaces(registry types.InterfaceRegistry){
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreateGreet{})
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

 
var (
amino = codec.NewLegacyAmino()
ModuleCdc = codec.NewAminoCodec(amino)
)
```

### Define our NewMsgCreateGreet type
Now that we have set up our codec & keys files lets add some code for creating a new greeting, create a file ```x/greet/types/message_greet.go``` and add the following code:

```
package  types

import (
	sdk  "github.com/cosmos/cosmos-sdk/types"
	sdkerrors  "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg =  &MsgCreateGreet{}

// this is our "constructor" for a new greeting 
func  NewMsgCreateGreet(owner string, message string) *MsgCreateGreet {
		return  &MsgCreateGreet{
		Owner: owner,
		Message: message,
	}
}

  
 // returns our RouterKey which is "greet"
func (msg MsgCreateGreet) Route() string {
	return RouterKey
}

  
// returns the type of our message which is "CreateGreet"
func (msg MsgCreateGreet) Type() string {
	return  "CreateGreet"
}

  
// get the signer of msg
func (msg MsgCreateGreet) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err !=  nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

  
// marshals the msg
func (msg *MsgCreateGreet) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

  

// does basic msg validation 
func (msg MsgCreateGreet) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err !=  nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address", err)
	}
	if  len(msg.Message) ==  0 {
		return sdkerrors.Wrapf(sdkerrors.Error{ }, "must provide greeting message")
	}

	return  nil
}
```

### Setting up a create-greet command 

In order for users to interact with out new module, we will create a command to create a new greeting that will live in the following folder so let's create it and write down some code in  ```/x/greet/client/cli/tx.go```:

```
package  cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/kava-labs/kava/x/greet/types"
)

  

// this is the parent tx command for the greet module everytime we add a new command we will register it here
func  GetTxCmd() *cobra.Command {
	cmd :=  &cobra.Command{
	Use: types.ModuleName,
	Short: fmt.Sprintf("%s transactions subcommands", types.ModuleName),
	DisableFlagParsing: true,
	SuggestionsMinimumDistance: 2,
	RunE: client.ValidateCmd,
	}

	// add the create greet command
	cmd.AddCommand(CmdCreateGreet())
	return cmd
}

  
 
// build the create greet command function
func  CmdCreateGreet() *cobra.Command {
	cmd :=  &cobra.Command{
	Use: "create-greet [message]",
	Short: "Create a new greeting",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		message :=  string(args[0])
		clientCtx, err := client.GetClientTxContext(cmd)
		if err !=  nil {
		return err
		}

		// create the new mesage using the new constuctor we made
		msg := types.NewMsgCreateGreet(clientCtx.GetFromAddress().String(), string(message))
		// do a basic message validation using the reciever function we created on the MsgCreateGreet type
		if err := msg.ValidateBasic(); err !=  nil {
		return err
		}

		// GenerateOrBroadcastTxCLI will either generate and print and unsigned transaction or sign it and broadcast it returning an error upon failure.

		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},

	}

	// add flags tx for our command
	flags.AddTxFlagsToCmd(cmd)

	// return the configured command struct
	return cmd
}
```
### Setting up query commands 
```/x/greet/client/cli/query.go```:
```
package  cli

import (
"fmt"
"github.com/spf13/cobra"
"github.com/cosmos/cosmos-sdk/client"
"github.com/cosmos/cosmos-sdk/client/flags"
"context"
"github.com/kava-labs/kava/x/greet/types"
)
  
// this is the parent query command for the greet module everytime we add a new command we will register it here
func  GetQueryCmd(queryRoute string) *cobra.Command {
	// Group todos queries under a subcommand
	cmd :=  &cobra.Command{
	Use: types.ModuleName,
	Short: fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
	DisableFlagParsing: true,
	SuggestionsMinimumDistance: 2,
	RunE: client.ValidateCmd,
	}
	// register the list greet command
	cmd.AddCommand(CmdListGreet())
	// register the show greet command
	cmd.AddCommand(CmdShowGreet())
	// return the configured command
	return cmd
}

// build the list greet command function
	func  CmdListGreet() *cobra.Command {
	cmd :=  &cobra.Command{
	Use: "list-greetings",
	Short: "list all greetings",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err !=  nil {
			return err
		}
		pageReq, err := client.ReadPageRequest(cmd.Flags())
		if err !=  nil {
			return err
		}
		queryClient := types.NewQueryClient(clientCtx)
		params :=  &types.QueryAllGreetRequest{
		Pagination: pageReq,
		}
		res, err := queryClient.GreetAll(context.Background(), params)
		if err !=  nil {
			return err
		}
		return clientCtx.PrintProto(res)
		},
		}
		flags.AddQueryFlagsToCmd(cmd)
		return cmd
	}

// build the show greet command function
func  CmdShowGreet() *cobra.Command {
	cmd :=  &cobra.Command{
	Use: "show-greet [id]",
	Short: "shows a greeting",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err !=  nil {
			return err
		}
		queryClient := types.NewQueryClient(clientCtx)
		params :=  &types.QueryGetGreetRequest{
			Id: args[0],
		}
		res, err := queryClient.Greet(context.Background(), params)
		if err !=  nil {
			return err
		}

		return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
```


## Setting Up Keeper

```x/greet/keeper/keeper.go```

```
package  keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk  "github.com/cosmos/cosmos-sdk/types"
	paramtypes  "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/swap/types"
)

type (
	Keeper struct {
	cdc codec.Codec
	key sdk.StoreKey
	paramSubspace paramtypes.Subspace
	}
)

func  NewKeeper(cdc codec.Codec, key sdk.StoreKey, paramstore paramtypes.Subspace, accountKeeper types.AccountKeeper) *Keeper {
	return  &Keeper{
	cdc: cdc,
	key: key,
	paramSubspace: paramstore,
	}
}
```
Setting up the greet keeper
```x/greet/keeper/greet.go```: 

```
package  keeper

import (
	"strconv"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	 sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/greet/types"
)  

func (k Keeper) GetGreetCount(ctx sdk.Context) int64 {
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetCountKey))
	byteKey := types.KeyPrefix(types.GreetCountKey)
	bz := store.Get(byteKey)
	if bz ==  nil {
	return  0
	}
	count, err := strconv.ParseInt(string(bz), 10, 64)
	if err !=  nil {
	panic("cannot decode count")
	}
	return count
}

func (k Keeper) SetGreetCount(ctx sdk.Context, count int64){
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetCountKey))
	byteKey := types.KeyPrefix(types.GreetCountKey)
	bz := []byte(strconv.FormatInt(count, 10))
	store.Set(byteKey, bz)
}

func (k Keeper) CreateGreet(ctx sdk.Context, msg types.MsgCreateGreet){
	count := k.GetGreetCount(ctx)
	var greet = types.Greet{
	Id: strconv.FormatInt(count, 10),
	Owner: msg.Owner,
	Message: msg.Message,
	}
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	key := types.KeyPrefix(types.GreetKey + greet.Id)
	value := k.cdc.MustMarshal(&greet)
	store.Set(key, value)
	k.SetGreetCount(ctx, count +  1)
}

 
func (k Keeper) GetGreet(ctx sdk.Context, key string) types.Greet {
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	var Greet types.Greet
	k.cdc.Unmarshal(store.Get(types.KeyPrefix(types.GreetKey + key)), &Greet)
	return Greet
}

func (k Keeper) HasGreet(ctx sdk.Context, id string) bool {
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	return store.Has(types.KeyPrefix(types.GreetKey + id))
}


func (k Keeper) GetGreetOwner(ctx sdk.Context, key string) string{
	return k.GetGreet(ctx, key).Owner
}

func (k Keeper) GetAllGreet(ctx sdk.Context) (msgs []types.Greet){
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefix(types.GreetKey))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var msg types.Greet
		k.cdc.Unmarshal(iterator.Value(), &msg)
		msgs =  append(msgs, msg)
	}
	return
}
```

```x/greet/query_greet.go```:

```
package  keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk  "github.com/cosmos/cosmos-sdk/types"
	sdkerrors  "github.com/cosmos/cosmos-sdk/types/errors"
)

func  getGreet(ctx sdk.Context, id string, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	msg := keeper.GetGreet(ctx, id)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, msg)
	if err !=  nil {
		return  nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}
func  listGreet(ctx sdk.Context, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	msgs := keeper.GetAllGreet(ctx)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, msgs)
	if err !=  nil {
		return  nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, err
}
```


```x/greet/querier.go```:

```
package  keeper

  

import (

"github.com/cosmos/cosmos-sdk/codec"

sdk  "github.com/cosmos/cosmos-sdk/types"

sdkerrors  "github.com/cosmos/cosmos-sdk/types/errors"

abci  "github.com/tendermint/tendermint/abci/types"

"github.com/kava-labs/kava/x/greet/types"

)

  
  

func  NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
		return  func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryGetGreet:
		return  getGreet(ctx, path[1], k, legacyQuerierCdc)
		case types.QueryListGreet:
		return  listGreet(ctx, k, legacyQuerierCdc)
		default:
		return  nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}
	}
}
```

```x/greet/types/query.go```:

```
package  types

const (
	QueryGetGreet =  "show-greet"
	QueryListGreet =  "list-greetings"
)
```

```x/greet/keeper/grpc_query_greet.go```:

```
package  keeper

import (
"context"
"github.com/cosmos/cosmos-sdk/store/prefix"
sdk  "github.com/cosmos/cosmos-sdk/types"
"github.com/cosmos/cosmos-sdk/types/query"
"github.com/kava-labs/kava/x/greet/types"
"google.golang.org/grpc/codes"
"google.golang.org/grpc/status"
)


func (k Keeper) GreetAll(c context.Context, req *types.QueryAllGreetRequest) (*types.QueryAllGreetResponse, error) {
	var greetings []*types.Greet
	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.key)
	greetStore := prefix.NewStore(store, types.KeyPrefix(types.GreetKey))

	pageRes, err := query.Paginate(greetStore, req.Pagination, func(key, value []byte) error {
	var greet types.Greet
	if err := k.cdc.Unmarshal(value, &greet); err !=  nil {
		return err
	}
	greetings =  append(greetings, &greet)
	return  nil
	})
	if err !=  nil {
		return  nil, status.Error(codes.Internal, err.Error())
	}
	return  &types.QueryAllGreetResponse{Greetings: greetings, Pagination: pageRes}, nil
}

  

func (k Keeper) Greet(c context.Context, req *types.QueryGetGreetRequest) (*types.QueryGetGreetResponse, error) {
	if req ==  nil {
		return  nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	var greet types.Greet
	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	k.cdc.MustUnmarshal(store.Get(types.KeyPrefix(types.GreetKey + req.Id)), &greet)
	return &types.QueryGetGreetResponse{Greeting: &greet}, nil
}
```

```x/greet/handler.go```

```
package  greet

import (
	"fmt"
	 sdk  "github.com/cosmos/cosmos-sdk/types"
	 sdkerrors  "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/kava-labs/kava/x/greet/keeper"
)

func  NewHandler(k keeper.Keeper) sdk.Handler{
	return  func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case  *types.MsgCreateGreet:
		return  handleMsgCreateGreet(ctx, k, msg)
		default:
		errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
		return  nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func  handleMsgCreateGreet(ctx sdk.Context, k keeper.Keeper, msg *types.MsgCreateGreet) (*sdk.Result, error) {
	k.CreateGreet(ctx, *msg)
	return  &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
```

```x/greet/module.go```

```
package  greet

import (
	"context"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes  "github.com/cosmos/cosmos-sdk/codec/types"
	sdk  "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/kava-labs/kava/x/greet/client/cli"
	"github.com/kava-labs/kava/x/greet/keeper"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/kava-labs/kava/x/pricefeed/client/rest"
	"github.com/spf13/cobra"
	abci  "github.com/tendermint/tendermint/abci/types"
)

  

var (
	_ module.AppModule = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type  AppModuleBasic  struct{} 
 
// Name get module name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec register module codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

  

// DefaultGenesis default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	err := cdc.UnmarshalJSON(bz, &gs)
	if err !=  nil {
		return err
	}
	return gs.Validate()
}


// RegisterInterfaces implements InterfaceModule.RegisterInterfaces
func (a AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

  
func (a AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}

  

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the gov module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err !=  nil {
		panic(err)
	}
}

  

// GetTxCmd returns the root tx command for the swap module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

  

// GetQueryCmd returns no root query command for the swap module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey)
}

// AppModule app module type
type  AppModule  struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func  NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper: keeper,
	}
}

// Name module name
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// RegisterInvariants register module invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) { }

func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

  

func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

  
  

func (AppModule) ConsensusVersion() uint64 {
	return  1
}

  

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

  
  

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)
	InitGenesis(ctx, am.keeper, genState)
	return []abci.ValidatorUpdate{}
}

  

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs :=  ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(&gs)
}

  

func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {	}

// EndBlock module end-block
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
```


```x/greet/msg_server.go```:

```
package  greet

import (
	"context"
	"github.com/kava-labs/kava/x/greet/keeper"
	"github.com/kava-labs/kava/x/greet/types"
	sdk  "github.com/cosmos/cosmos-sdk/types"
)

  
  

type  msgServer  struct {
	keeper keeper.Keeper
}

func (m msgServer) CreateGreet(c context.Context, msg *types.MsgCreateGreet) (*types.MsgCreateGreetResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	m.keeper.CreateGreet(ctx, types.MsgCreateGreet{Owner: msg.Owner, Message: msg.Message})
	return  &types.MsgCreateGreetResponse{}, nil
}

func  NewMsgServerImpl(keeper keeper.Keeper) types.MsgServer {
	return  &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}
```