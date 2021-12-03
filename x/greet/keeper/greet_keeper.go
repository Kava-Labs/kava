package keeper

import (
	"context"
	"strconv"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/greet/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Keeper struct {
		cdc codec.Codec
		key sdk.StoreKey
		paramSubspace paramtypes.Subspace
}


func NewKeeper(c codec.Codec, k sdk.StoreKey, pss paramtypes.Subspace) Keeper{
	return Keeper{
		cdc: c,
		key: k,
		paramSubspace: pss,
	}
}


// get greet count for ids 
func (k Keeper) GetGreetCount(ctx sdk.Context) int64 {
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetCountKey))
	byteKey := types.KeyPrefix(types.GreetCountKey)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0
	}

	count, err := strconv.ParseInt(string(bz), 10, 64)
	if err != nil {
		panic("cannot decode count")
	}

	return count
}

func (k Keeper) SetGreetCount(ctx sdk.Context, count int64){
    store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetCountKey))
	key := types.KeyPrefix(types.GreetCountKey)
	value := []byte(strconv.FormatInt(count, 10))
	store.Set(key, value)
}


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

	k.SetGreetCount(ctx, count + 1)
}

func (k Keeper) GetGreeting(ctx sdk.Context, key string) types.Greet {
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	var Greet types.Greet

	k.cdc.Unmarshal(store.Get(types.KeyPrefix(types.GreetKey + key)), &Greet)
	return Greet
}

func (k Keeper) HasGreet(ctx sdk.Context, id string) bool {
	store :=  prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	return store.Has(types.KeyPrefix(types.GreetKey + id))
}

func (k Keeper) GetGreetOwner(ctx sdk.Context, key string) string {
	return k.GetGreeting(ctx, key).Owner
}

func (k Keeper) GetAllGreetings(ctx sdk.Context) (msgs []types.Greet){
	store := prefix.NewStore(ctx.KVStore(k.key), types.KeyPrefix(types.GreetKey))
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefix(types.GreetKey))

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var msg types.Greet
		k.cdc.Unmarshal(iterator.Value(), &msg)
		msgs = append(msgs, msg)
	}

	return 
}



// GRPC QUERY 
func (k Keeper) GreetAll(c context.Context, req *types.QueryAllGreetRequest) (*types.QueryAllGreetResponse, error){
	ctx := sdk.UnwrapSDKContext(c)
	var greetings []*types.Greet

	for _, g := range k.GetAllGreetings(ctx) {
		var greeting = &g 
		greetings = append(greetings,greeting)
	}

	return &types.QueryAllGreetResponse{Greetings: greetings, Pagination: nil}, nil
}


func (k Keeper) Greet(c context.Context, req *types.QueryGetGreetRequest) (*types.QueryGetGreetResponse, error){
	sdk.UnwrapSDKContext(c)
	var greeting = k.GetGreeting(sdk.UnwrapSDKContext(c), req.Id)
	return &types.QueryGetGreetResponse{Greeting: &greeting}, nil 
}


// LEGACY QUERIER 
func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryGetGreeting:
				var getGreetRequest types.QueryGetGreetRequest 
				err := legacyQuerierCdc.UnmarshalJSON(req.Data, &getGreetRequest)
				if err != nil {
					return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
				}
				
				val := k.GetGreeting(ctx, getGreetRequest.GetId())
				bz, err := legacyQuerierCdc.MarshalJSON(val) 
				if err != nil {
					return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
				}
			return bz, nil 
		case types.QueryListGreetings: 
				val := k.GetAllGreetings(ctx)
				bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, val)
				if err != nil {
					return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
				}
			return bz, nil 
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknow request at %s query endpoint", types.ModuleName)
		}
	}
}

