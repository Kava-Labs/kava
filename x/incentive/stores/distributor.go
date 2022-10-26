package stores

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/distributor"
	"github.com/kava-labs/kava/x/incentive/types"
)

var _ distributor.DistributorStore = distributorStore{}

func NewDistributorStore(cdc codec.Codec, key sdk.StoreKey, reward types.RewardType) distributor.DistributorStore {
	return distributorStore{
		cdc:    cdc,
		key:    key,
		reward: reward,
	}
}

type distributorStore struct {
	cdc    codec.Codec
	key    sdk.StoreKey
	reward types.RewardType
}

func (s distributorStore) GetClaim(ctx sdk.Context, addr sdk.AccAddress) (types.Claim, bool) {
	store := prefix.NewStore(ctx.KVStore(s.key), types.ClaimsKeyPrefix)
	bz := store.Get(types.NewClaimKey(s.reward, addr))
	if bz == nil {
		return types.Claim{}, false
	}
	var c types.Claim
	s.cdc.MustUnmarshal(bz, &c)
	return c, true
}

// SetClaim sets the claim in the store corresponding to the input address.
func (s distributorStore) SetClaim(ctx sdk.Context, c types.Claim) {
	store := prefix.NewStore(ctx.KVStore(s.key), types.ClaimsKeyPrefix)
	bz := s.cdc.MustMarshal(&c)
	store.Set(types.NewClaimKey(s.reward, c.Owner), bz)
}

func (s distributorStore) GetGlobalIndexes(ctx sdk.Context, rewardID string) (types.RewardIndexes, bool) {
	store := prefix.NewStore(ctx.KVStore(s.key), types.GlobalIndexesKeyPrefix)
	bz := store.Get(types.NewGlobalIndexesKey(s.reward, rewardID))
	if bz == nil {
		return types.RewardIndexes{}, false
	}
	var proto types.RewardIndexesProto
	s.cdc.MustUnmarshal(bz, &proto)
	return proto.RewardIndexes, true
}

func (s distributorStore) SetGlobalIndexes(ctx sdk.Context, rewardID string, indexes types.RewardIndexes) {
	if len(rewardID) == 0 {
		panic("invalid reward ID")
	}

	store := prefix.NewStore(ctx.KVStore(s.key), types.GlobalIndexesKeyPrefix)
	bz := s.cdc.MustMarshal(&types.RewardIndexesProto{
		RewardIndexes: indexes,
	})
	store.Set(types.NewGlobalIndexesKey(s.reward, rewardID), bz)
}

// IterateGlobalIndexes iterates over all swap reward index objects in the store and preforms a callback function
func (s distributorStore) IterateGlobalIndexes(ctx sdk.Context, cb func(poolID string, indexes types.RewardIndexes) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(s.key), types.GlobalIndexesKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{byte(s.reward)}) // TODO
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var proto types.RewardIndexesProto
		s.cdc.MustUnmarshal(iterator.Value(), &proto)
		if cb(string(iterator.Key()[1:]), proto.RewardIndexes) {
			break
		}
	}
}
