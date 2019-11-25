package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/kava-labs/kava/x/auction/types"
	"github.com/tendermint/tendermint/crypto"
)

func setUpMockApp() (*mock.App, Keeper, []sdk.AccAddress, []crypto.PrivKey) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	types.RegisterCodec(mapp.Cdc)

	// Create keepers
	keyAuction := sdk.NewKVStoreKey("auction")
	blacklistedAddrs := make(map[string]bool)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)
	auctionKeeper := NewKeeper(mapp.Cdc, bankKeeper, keyAuction, mapp.ParamsKeeper.Subspace(types.DefaultParamspace))

	// Mount and load the stores
	err := mapp.CompleteSetup(keyAuction)
	if err != nil {
		panic("mock app setup failed")
	}

	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(10, sdk.NewCoins(sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 100)))
	mock.SetGenesis(mapp, genAccs)

	return mapp, auctionKeeper, addrs, privKeys
}
