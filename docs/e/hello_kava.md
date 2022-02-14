#  Building a Module 

In this tutorial we will be going over building a module in Kava to show how easy it is to build on top of the Kava ecosystem. This module will be simple in nature but will show how to set up and connect a module to Kava and can be used as a starting point for more complex modules. 



	
## Set up
```
this tutorial is based on the v44 cosmos version which Kava is currently migrating to, to follow this tutorial clone the kava repo but ensure you 
clone the upgrade-v44 branch as the master branch is currently on v39 & ensure you have kvtool, docker & go installed on your machine. 
git clone -b upgrade-v44 https://github.com/Kava-Labs/kava.git 
```
## Defining Protocol Buffer Types 

The first step in building a new Kava Module is to define our Module's types. To do that we use Protocol Buffers which is a used for serializing structured data and generating code for multiple target languages, Protocol Buffers are also smaller than JSON & XML so sending data around the network will be less expensive. [Learn More](https://developers.google.com/protocol-buffers). 

Our Protobuf files will all live in ```proto/kava``` directory.  we will create a new directory with the new module ```greet``` and add the following files in the ```proto/greet/v1beta1/``` directory
```
genesis.proto
greet.proto
query.proto
tx.proto
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
Here we are saying that we have a Greet type that will have an owner, an id and a message that will contain the greet string. Once we have that defined we are ready to set up a way to create this greet message and query it.  

### Creating a new Greeting 
Inside the ```proto/greet/v1beta1/tx.proto``` file lets define our Msg Type: 
```
syntax = "proto3";
package  kava.greet.v1beta1;
import  "gogoproto/gogo.proto";
import  "cosmos_proto/cosmos.proto";
option  go_package = "github.com/kava-labs/kava/x/greet/types";

service  Msg {
	rpc  CreateGreet(MsgCreateGreet) returns (MsgCreateGreetResponse);
}

message  MsgCreateGreet {
string message = 1;
string owner = 2;
}
message  MsgCreateGreetResponse {}
```
Now that we have defined how to create a new Greeting let's finish up by setting up our queries to view a specific greeting or all of them.

One thing to note here is that any state changing actions are transactions and for that reason we put them in our ```tx.proto``` files, we essentially said we are creating a new state changing message & defined the types for that message in our proto file, we will later add clients to trigger state change, which in our case will be adding a new message to our chain. 

### Querying Greetings 
Code inside the ```proto/greet/v1beta1/query.proto``` : 
```
syntax = "proto3";

package  kava.greet.v1beta1;
option  go_package = "github.com/kava-labs/kava/x/greet/types";

import  "gogoproto/gogo.proto";
import  "google/api/annotations.proto";
import  "cosmos/base/query/v1beta1/pagination.proto";
import  "cosmos_proto/cosmos.proto";
import  "kava/greet/v1beta1/greet.proto";

service  Query {
	rpc  Greet(QueryGetGreetRequest) returns (QueryGetGreetResponse) {
	option  (google.api.http).get = "/kava/greet/v1beta1/greetings/{id}";
	}
	rpc  GreetAll(QueryAllGreetRequest) returns (QueryAllGreetResponse) {
	option  (google.api.http).get = "/kava/swap/v1beta1/greetings";
	}
}

 
message  QueryGetGreetRequest {
string id = 1;
}

message  QueryGetGreetResponse {
Greet greeting = 1;
}

message  QueryAllGreetRequest {
cosmos.base.query.v1beta1.PageRequest pagination = 1;
}

message  QueryAllGreetResponse {
repeated  Greet greetings = 1;
cosmos.base.query.v1beta1.PageResponse pagination = 2;
}
```
Our ```query.proto``` now contains the types for our queries, we have defined a request type  & a response type and those types will be returned once we trigger a query through the CLI, REST API, or Grpc. The response will follow the same structure regardless of the type of client initiating the request. 

We defined our query, tx, and greet proto files we finally need to set up the genesis file and then we are ready to generate these types. In the genesis file we will create a minimal ```genesis.proto``` for this tutorial to keep things simple. 
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

## Developing Our Greet Module
we have successfully set up our Proto files & generated them, we now have a ```x/greet``` directory generated, this is where we will write our module's code. For starters we will define our module's types in a new file inside ```x/greet/types/greet.go```. 

### Setting up constants & importing packages 
Let's set up some basic constants for our module to help with routing, & fetching items from our store. 
```
package  types

import (
	"fmt"
	"strings"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk  "github.com/cosmos/cosmos-sdk/types"
	sdkerrors  "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

  

// defined our module's constants such as name, routerkey 
// and prefixes for getting items from the store
const (
	ModuleName =  "greet"
	StoreKey = ModuleName 
	RouterKey = ModuleName  
	QuerierRoute = ModuleName  
	GreetKey =  "greet-value-"  // used for getting a greeting from our store
	GreetCountKey =  "greet-count-"  // used for getting count from out store
	QueryGetGreeting =  "get-greeting"  // used for legacy querier routing
	QueryListGreetings =  "list-greetings"// used for legacy querier routing
)
// heler function simply returns []byte out of a prefix string
func  KeyPrefix(p string) []byte {
	return []byte(p)
}

// returns default genesis state
func  DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// validates genesis state
func (gs GenesisState) Validate() error {
	return  nil
}
```

### Setting up our Msg for creating a new greeting 
Our ```MsgCreateGreet``` struct was created when we generated our Proto Types, we now need to use that struct to implement the ```sdk.Msg``` interface such that we can create new greetings. the first thing we will do is defined an unnamed variable with the ```_``` syntax and have it implement the ```sdk.Msg``` type. This will help us catch unimplemented functions and guide us with syntax highlighting. 

```
// MsgCreateGreet we defined it here to get type checking 
//to make sure we are immplementing it correctly
var _ sdk.Msg =  &MsgCreateGreet{}

  
// constructor for creating a new greeting
func  NewMsgCreateGreet(owner string, message string) *MsgCreateGreet{
	return  &MsgCreateGreet{
	Owner: owner,
	Message: message,
	}
}
// does a quick stateless validation on our new greeting 
func (m *MsgCreateGreet) ValidateBasic() error {
	// ensures address is valid 
	if _, err := sdk.AccAddressFromBech32(m.Owner); err !=  nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address", err)
	}
	// ensures the greeting is not empty
	if  len(strings.TrimSpace(m.Message)) ==  0 {
		return fmt.Errorf("must provide a greeting message")
	}
	return  nil
}

// gets the signer of the new message which will be the owner of the greeting 
func (m *MsgCreateGreet) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner);
	if err !=  nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}
```

### Registering our Codec & interfaces 
now that we have our  ```MsgCreateGreet``` implement the ```sdk.Msg``` interface let's register our codec for marshaling/unmarshaling our greeting we will register both the deprecated legacy amino and the new Interface registry. 
```
// registers the marshal/unmarsahl for greating a new greeting for our legacy amino codec
func  RegisterLegacyAminoCodec(cdc *codec.LegacyAmino){
	cdc.RegisterConcrete(&MsgCreateGreet{}, "greet/CreateGreet", nil)
}

// registers a module's interface types and their concrete implementations as proto.Message.
func  RegisterInterfaces(registry types.InterfaceRegistry){
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreateGreet{})
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var amino = codec.NewLegacyAmino()
var ModuleCdc = codec.NewAminoCodec(amino)
```

### Setting up a basic Keeper 
we have finished up setting up our types, now it's time to implement our greet module's keeper, lets do that in a new folder & package named keeper, create ```x/greet/keeper/greet_keeper.go``` .

### Setting up the Keeper Struct & imports 
keepers are an abstraction over the state defined by a module, every module would have a keeper which would be used to access the state of that module, or if given access a keeper can also use other module's keepers by providing reference to the other module's keeper. 
```
package  keeper

import (
	"context"
	"strconv"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk  "github.com/cosmos/cosmos-sdk/types"
	abci  "github.com/tendermint/tendermint/abci/types"
	"github.com/kava-labs/kava/x/greet/types"
	sdkerrors  "github.com/cosmos/cosmos-sdk/types/errors"
)

type  Keeper  struct {
	cdc codec.Codec // used to marshall and unmarshall structs from & to []byte
	key sdk.StoreKey // grant access to the store
}

// our constructor for creating a new Keeper for this module
func  NewKeeper(c codec.Codec, k sdk.StoreKey) Keeper {
	return Keeper{
	cdc: c,
	key: k,
	}
}
```

### Wiring up our methods for handling new transactions & queries 
Now that we have our Keeper Struct written, let's create some receiver functions on our keeper to handle adding a new greeting & looking up a greeting. 
```
// get greet count will be used for setting an Id when a new greeting is created
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

// sets the greet count
func (k Keeper) SetGreetCount(ctx sdk.Context, count int64){
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetCountKey))
	key := types.KeyPrefix(types.GreetCountKey)
	value := []byte(strconv.FormatInt(count, 10))
	store.Set(key, value)
}

// creates a new greeting
func (k Keeper) CreateGreet(ctx sdk.Context, m types.MsgCreateGreet){
	count := k.GetGreetCount(ctx)
	greet := types.Greet{
	Id: strconv.FormatInt(count, 10),
	Owner: m.Owner,
	Message: m.Message,
	}
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	key := types.KeyPrefix(types.GreetKey + greet.Id)
	value := k.cdc.MustMarshal(&greet)
	store.Set(key, value)
	k.SetGreetCount(ctx, count +  1)
}

// gets a greeting from the store
func (k Keeper) GetGreeting(ctx sdk.Context, key string) types.Greet {
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	var Greet types.Greet
	k.cdc.Unmarshal(store.Get(types.KeyPrefix(types.GreetKey + key)), &Greet)
	return Greet
}

// checks if a greeting exists by an id
func (k Keeper) HasGreet(ctx sdk.Context, id string) bool {
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	return store.Has(types.KeyPrefix(types.GreetKey + id))
}

// gets the owner of a greeting
func (k Keeper) GetGreetOwner(ctx sdk.Context, key string) string {
	return k.GetGreeting(ctx, key).Owner
}

// gets a list of all greetings in the store
func (k Keeper) GetAllGreetings(ctx sdk.Context) (msgs []types.Greet){
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
### Handling queries 

We have added methods for interacting with greetings such as creating or reading them, now let's set up our two query services so we can route them to the correct method, we will set up our legacy Querier & gRPC querier below the methods we defined above on our keeper. 

```
func (k Keeper) GreetAll(c context.Context, req *types.QueryAllGreetRequest) (*types.QueryAllGreetResponse, error){
	ctx := sdk.UnwrapSDKContext(c)
	var greetings []*types.Greet
	for _, g :=  range k.GetAllGreetings(ctx) {
		var greeting =  &g
		greetings =  append(greetings,greeting)
	}
	return  &types.QueryAllGreetResponse{Greetings: greetings, Pagination: nil}, nil
}

func (k Keeper) Greet(c context.Context, req *types.QueryGetGreetRequest) (*types.QueryGetGreetResponse, error){
	sdk.UnwrapSDKContext(c)
	var greeting = k.GetGreeting(sdk.UnwrapSDKContext(c), req.Id)
	return  &types.QueryGetGreetResponse{Greeting: &greeting}, nil
}

  
  

// LEGACY QUERIER will be deperacted but for the sake of competeness this is how to set it up
func  NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return  func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	switch path[0] {
		case types.QueryGetGreeting:
			var getGreetRequest types.QueryGetGreetRequest
			err := legacyQuerierCdc.UnmarshalJSON(req.Data, &getGreetRequest)
			if err !=  nil {
				return  nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
			}
			val := k.GetGreeting(ctx, getGreetRequest.GetId())
			bz, err := legacyQuerierCdc.MarshalJSON(val)
			if err !=  nil {
				return  nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
			}
			return bz, nil
		
		case types.QueryListGreetings:
			val := k.GetAllGreetings(ctx)
			bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, val)
			if err !=  nil {
				return  nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
			}
			return bz, nil
		default:
			return  nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknow request at %s query endpoint", types.ModuleName)
		}
	}
}
```

### Setting up a command to create a new greeting 
let's set up a way for clients to submit a new greeting & query existing greetings, we can do that with a CLI, REST, & gRPC clients. for this tutorial we will focus on setting up our CLI client.  create ```x/greet/client/cli/tx.go```. 

here We will define a command to create a new greeting:
```
package  cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/spf13/cobra"
)

func  GetTxCmd() *cobra.Command {
	cmd :=  &cobra.Command{
	Use: types.ModuleName,
	Short: fmt.Sprintf("%s transactions subcommands", types.ModuleName),
	DisableFlagParsing: true,
	SuggestionsMinimumDistance: 2,
	RunE: client.ValidateCmd,
	}
	cmd.AddCommand(CmdCreateGreeting())
	return cmd
}

 
func  CmdCreateGreeting() *cobra.Command {
	cmd:=  &cobra.Command{
	Use: "create-greeting [message]",
	Short: "creates a new greetings",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
			message :=  string(args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err !=  nil {
				return err
			}
			msg := types.NewMsgCreateGreet(clientCtx.GetFromAddress().String(), string(message))
			if err := msg.ValidateBasic(); err !=  nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
```

### Querying greetings 
We will now set up two different commands for querying, one will be to list all greetings & the other will be to get a greeting by it's id. inside ```x/greet/cli/query.go```:

```
package  cli

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/spf13/cobra"
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

	cmd.AddCommand(CmdListGreetings())
	cmd.AddCommand(CmdShowGreeting())
	return cmd
}

// build the list greet command function
func  CmdListGreetings() *cobra.Command {
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
func  CmdShowGreeting() *cobra.Command {
	cmd :=  &cobra.Command{
	Use: "get-greeting [id]",
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

### Setting up our Module's package 

Now that we have all the basic functionality set up for our greet module, let's bring it all together and get our module ready to be used & tested, create a new file ```x/greet/module.go```. 

Here we will start by implementing our ```AppModuleBasic```  && ```AppModule``` interfaces. 

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
	"github.com/spf13/cobra"
	abci  "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

/*
The AppModuleBasic interface defines the independent methods modules need to implement
it follows this interface below
type AppModuleBasic interface {
	Name() string
	RegisterLegacyAminoCodec(*codec.LegacyAmino)
	RegisterInterfaces(codectypes.InterfaceRegistry)
	DefaultGenesis(codec.JSONMarshaler) json.RawMessage
	ValidateGenesis(codec.JSONMarshaler, client.TxEncodingConfig, json.RawMessage) error
	// client functionality
	RegisterRESTRoutes(client.Context, *mux.Router)
	RegisterGRPCRoutes(client.Context, *runtime.ServeMux)
	GetTxCmd() *cobra.Command
	GetQueryCmd() *cobra.Command
}
*/

type  AppModuleBasic  struct{}

// Returns the name of the module as a string
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&gs)
}


func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return  nil
}
// Registers the amino codec for the module, which is used to marshal
// and unmarshal structs to/from []byte in order to persist them in the module's KVStore.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino){
	types.RegisterLegacyAminoCodec(cdc)
}
// Registers a module's interface types and their concrete implementations as proto.Message
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}
// Registers gRPC routes for the module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err !=  nil {
		panic(err)
	}
}
// Registers the REST routes for the module. These routes will be used to map REST request to the module in order to process them
func (a AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) { }

// Returns the root Tx command for the module. The subcommands of this root command are used by end-users
// to generate new transactions containing messages defined in the module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}
// Return the root query command for the module. The subcommands of this root command are used by end-users
// to generate new queries to the subset of the state defined by the module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.StoreKey)
}

// -------------------------------------APPMODULE BELOW------------------------------------------------- //

  

/*
The AppModule interface defines the inter-dependent methods that modules need to implement
follows the interface below
	type AppModule interface {
		AppModuleGenesis
		// registers
		RegisterInvariants(sdk.InvariantRegistry)
		// routes
		Route() sdk.Route
		// Deprecated: use RegisterServices
		QuerierRoute() string
		// Deprecated: use RegisterServices
		LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier
		// RegisterServices allows a module to register services
		RegisterServices(Configurator)
		// ABCI
		BeginBlock(sdk.Context, abci.RequestBeginBlock)
		EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate
	}
*/
type  AppModule  struct{
	AppModuleBasic
	keeper keeper.Keeper
}
// constructor
func  NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper: keeper,
	}
}
// Returns the route for messages to be routed to the module by BaseApp.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// registers the invariants of the module. If an invariant deviates from its predicted value,
// the InvariantRegistry triggers appropriate logic (most often the chain will be halted).
func (AppModule) RegisterInvariants(ir sdk.InvariantRegistry) { }

// Returns the route for messages to be routed to the module by BaseApp.
func (AppModule) Route() sdk.Route {
	return sdk.Route{}
}
  

// Returns the name of the module's query route, for queries to be routes to the module by BaseApp.deprecated
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// Returns a querier given the query path, in order to process the query.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

func (AppModule) ConsensusVersion() uint64 {
	return  1
}

// Allows a module to register services.
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

func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) { }
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

  

// ----------------------------------MSGSERVER REGISTER------------------------//
var _ types.MsgServer = msgServer{}
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
```

### Hooking up our module inside App.go

inside ```app/app.go``` start off importing the greet module, it's types & keeper packages and add them to the following places: 

1. ```module.NewBasicManager()``` add ```greet.AppModuleBasic{}```
2.  ```type App struct {}```  add ```greetkeeper.Keeper```
3.  ```sdk.NewKVStoreKeys()``` inside ```NewApp``` func  add ```greettypes.StoreKey```
4.  inside ```NewApp``` func add ```app.greetKeeper = greetKeeper.NewKeeper()``` and add arguments ```appCodec``` & ```keys[greettypes.StoreKey]```
5. inside ```NewApp``` find where we define ```app.mm``` & add ```greet.NewAppModule(app.greetKeeper),```
6. finally add the greet module's name to ```SetOrderBeginBlockers```, ```SetOrderEndBlockers``` && ```SetOrderInitGenesis```

## Testing our new Module 

1. inside the root of our directory run ```docker build -t kava/kava:tutorial-demo .```
2. find the directory for ```kvtool``` and open in your favorite code editor 
3. run ```kvtool testnet gen-config kava --kava.configTemplate upgrade-v44``` which will create a bunch of files inside ```full_configs/generated```
4. open up the two ```docker-compose.yaml``` files the one inside ```generated``` & the one inside ```generated/kava``` and change the image to point to ```kava/kava:tutorial-demo``` this will point to the local image we just built 
5. change into the ```full_configs/generated``` directory and run ```docker compose up -d```
6. now run ```docker compose exec kavanode bash``` to bash into our ```kava``` cli inside the running container

We should now have access to our greet commands that we defined first we will test creating a new greeting, for that we will run the following command: 

```kava tx greet create-greeting "hello world from kava chain" --from whale```

now let's test to see if the greeting message is able to be queried: 

```kava q greet list-greetings```

We should see something like this below: 

```
greetings:
- id: "0"
  message: hello world from kava chain
  owner: kava173w2zz287s36ewnnkf4mjansnthnnsz7rtrxqc
pagination: null
```

Now let's test if we can query the greeting by it's id which in our case will be ```"0"```, run the following: 

```kava q greet get-greeting 0```

We should see:
```
greeting:
  id: "0"
  message: hello world from kava chain
  owner: kava173w2zz287s36ewnnkf4mjansnthnnsz7rtrxqc
```

