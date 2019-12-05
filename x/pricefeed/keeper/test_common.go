package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

type testHelper struct {
	mApp     *mock.App
	keeper   Keeper
	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

func getMockApp(t *testing.T, numGenAccs int, genState types.GenesisState, genAccs []authexported.Account) testHelper {
	mApp := mock.NewApp()
	types.RegisterCodec(mApp.Cdc)
	keyPricefeed := sdk.NewKVStoreKey(types.StoreKey)
	pk := mApp.ParamsKeeper
	keeper := NewKeeper(mApp.Cdc, keyPricefeed, pk.Subspace(types.DefaultParamspace), types.DefaultCodespace)

	require.NoError(t, mApp.CompleteSetup(keyPricefeed))

	valTokens := sdk.TokensFromConsensusPower(42)
	var (
		addrs    []sdk.AccAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens)))
	}

	mock.SetGenesis(mApp, genAccs)
	return testHelper{mApp, keeper, addrs, pubKeys, privKeys}
}
