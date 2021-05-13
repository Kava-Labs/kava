package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// NewTestContext sets up a basic context with an in-memory db
func NewTestContext(requiredStoreKeys ...sdk.StoreKey) sdk.Context {
	memDB := db.NewMemDB()
	cms := store.NewCommitMultiStore(memDB)

	for _, key := range requiredStoreKeys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	cms.LoadLatestVersion()

	return sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
}

// unitTester is a wrapper around suite.Suite, with common functionality for keeper unit tests.
// It can be embedded in structs the same way as suite.Suite.
type unitTester struct {
	suite.Suite
	keeper keeper.Keeper
	ctx    sdk.Context
}

func (suite *unitTester) SetupTest() {
	incentiveStoreKey := sdk.NewKVStoreKey(types.StoreKey)
	suite.keeper = suite.setupKeeper(incentiveStoreKey)
	suite.ctx = NewTestContext(incentiveStoreKey)
}

func (suite *unitTester) TearDownTest() {
	suite.keeper = keeper.Keeper{}
	suite.ctx = sdk.Context{}
}

func (suite *unitTester) setupKeeper(incentiveStoreKey sdk.StoreKey) keeper.Keeper {
	cdc := app.MakeCodec()
	// TODO The param store key needs to be initialized in the multistore if params are to be Get/Set.
	paramSubspace := params.NewKeeper(
		cdc,
		sdk.NewKVStoreKey(params.StoreKey),
		sdk.NewTransientStoreKey(params.StoreKey),
	).Subspace(incentive.DefaultParamspace)

	return keeper.NewKeeper(cdc, incentiveStoreKey, paramSubspace, nil, nil, nil, nil, nil)
}

func (suite *unitTester) storeGlobalIndexes(indexes types.MultiRewardIndexes) {
	for _, i := range indexes {
		suite.keeper.SetHardBorrowRewardIndexes(suite.ctx, i.CollateralType, i.RewardIndexes)
	}
}

func (suite *unitTester) storeClaim(claim types.HardLiquidityProviderClaim) {
	suite.keeper.SetHardLiquidityProviderClaim(suite.ctx, claim)
}

func arbitraryCoins() sdk.Coins {
	return cs(c("btcb", 1))
}

func arbitraryCoinsWithDenoms(denom ...string) sdk.Coins {
	const arbitraryAmount = 1 // must be > 0 as sdk.Coins type only stores positive amounts
	coins := sdk.NewCoins()
	for _, d := range denom {
		coins = coins.Add(sdk.NewInt64Coin(d, arbitraryAmount))
	}
	return coins
}

func arbitraryAddress() sdk.AccAddress {
	_, addresses := app.GeneratePrivKeyAddressPairs(1)
	return addresses[0]
}

var nonEmptyMultiRewardIndexes = types.MultiRewardIndexes{
	{
		CollateralType: "bnb",
		RewardIndexes: types.RewardIndexes{
			{
				CollateralType: "hard",
				RewardFactor:   d("0.02"),
			},
			{
				CollateralType: "ukava",
				RewardFactor:   d("0.04"),
			},
		},
	},
	{
		CollateralType: "btcb",
		RewardIndexes: types.RewardIndexes{
			{
				CollateralType: "hard",
				RewardFactor:   d("0.2"),
			},
			{
				CollateralType: "ukava",
				RewardFactor:   d("0.4"),
			},
		},
	},
}

func extractCollateralTypes(indexes types.MultiRewardIndexes) []string {
	var denoms []string
	for _, ri := range indexes {
		denoms = append(denoms, ri.CollateralType)
	}
	return denoms
}

func increaseAllRewardFactors(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	increasedIndexes := make(types.MultiRewardIndexes, len(indexes))
	copy(increasedIndexes, indexes)

	for i := range increasedIndexes {
		increasedIndexes[i].RewardIndexes = increaseRewardFactors(increasedIndexes[i].RewardIndexes)
	}
	return increasedIndexes
}

func increaseRewardFactors(indexes types.RewardIndexes) types.RewardIndexes {
	increasedIndexes := make(types.RewardIndexes, len(indexes))
	copy(increasedIndexes, indexes)

	for i := range increasedIndexes {
		increasedIndexes[i].RewardFactor = increasedIndexes[i].RewardFactor.MulInt64(2)
	}
	return increasedIndexes
}

func appendUniqueMultiRewardIndex(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	const uniqueDenom = "uniquedenom"

	for _, mri := range indexes {
		if mri.CollateralType == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique multi reward index with denom '%s', but denom already existed", uniqueDenom))
		}
	}

	return append(indexes, types.NewMultiRewardIndex(
		uniqueDenom,
		types.RewardIndexes{
			{
				CollateralType: "hard",
				RewardFactor:   d("0.02"),
			},
			{
				CollateralType: "ukava",
				RewardFactor:   d("0.04"),
			},
		},
	),
	)
}

func appendUniqueEmptyMultiRewardIndex(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	const uniqueDenom = "uniquedenom"

	for _, mri := range indexes {
		if mri.CollateralType == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique multi reward index with denom '%s', but denom already existed", uniqueDenom))
		}
	}

	return append(indexes, types.NewMultiRewardIndex(uniqueDenom, nil))
}

func appendUniqueRewardIndexToFirstItem(indexes types.MultiRewardIndexes) types.MultiRewardIndexes {
	newIndexes := make(types.MultiRewardIndexes, len(indexes))
	copy(newIndexes, indexes)

	newIndexes[0].RewardIndexes = appendUniqueRewardIndex(newIndexes[0].RewardIndexes)
	return newIndexes
}

func appendUniqueRewardIndex(indexes types.RewardIndexes) types.RewardIndexes {
	const uniqueDenom = "uniquereward"

	for _, mri := range indexes {
		if mri.CollateralType == uniqueDenom {
			panic(fmt.Sprintf("tried to add unique reward index with denom '%s', but denom already existed", uniqueDenom))
		}
	}

	return append(
		indexes,
		types.NewRewardIndex(uniqueDenom, d("0.02")),
	)
}
