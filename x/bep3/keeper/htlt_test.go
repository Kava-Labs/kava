package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type HtltTestSuite struct {
	suite.Suite

	keeper keeper.Keeper
	app    app.TestApp
	ctx    sdk.Context
}

var (
	coinsSingle  = sdk.Coins{sdk.Coin{Denom: "bnb", Amount: sdk.NewInt(50000)}}
	coinsZero    = sdk.Coins{sdk.Coin{}}
	timestamp1   = int64(6655443322)
	timestamp2   = int64(7766554433)
	timestamp3   = int64(8877665544)
	timestamp4   = int64(9988776655)
	rnh1         = types.BytesToHexEncodedString(types.CalculateRandomHash([]byte{15}, timestamp1))
	rnh2         = types.BytesToHexEncodedString(types.CalculateRandomHash([]byte{72}, timestamp2))
	rnh3         = types.BytesToHexEncodedString(types.CalculateRandomHash([]byte{119}, timestamp3))
	rnh4         = types.BytesToHexEncodedString(types.CalculateRandomHash([]byte{154}, timestamp4))
	binanceAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest3"))),
		sdk.AccAddress(crypto.AddressHash([]byte("BinanceTest4"))),
	}
	kavaAddrs = []sdk.AccAddress{
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest2"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest3"))),
		sdk.AccAddress(crypto.AddressHash([]byte("KavaTest4"))),
	}
)

func (suite *HtltTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	keeper := tApp.GetBep3Keeper()
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	return
}

func (suite *HtltTestSuite) TestGetSetHtlt() {
	htlt := types.NewHTLT(binanceAddrs[0], kavaAddrs[0], "", "", rnh1, timestamp1, coinsSingle, "bnb50000", 80000, false)
	swapID, err := types.CalculateSwapID(htlt.RandomNumberHash, htlt.From, htlt.SenderOtherChain)
	suite.NoError(err)
	suite.keeper.SetHTLT(suite.ctx, htlt, swapID)

	h, found := suite.keeper.GetHTLT(suite.ctx, swapID)
	suite.True(found)
	suite.Equal(htlt, h)

	fakeSwapID, err := types.CalculateSwapID(htlt.RandomNumberHash, kavaAddrs[1], "otheraddress")
	suite.NoError(err)
	_, found = suite.keeper.GetHTLT(suite.ctx, fakeSwapID)
	suite.False(found)

	suite.keeper.DeleteHTLT(suite.ctx, swapID)
	_, found = suite.keeper.GetHTLT(suite.ctx, swapID)
	suite.False(found)
}

func (suite *HtltTestSuite) TestAddHtlt() {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.app.GetAccountKeeper()
	acc := ak.NewAccountWithAddress(suite.ctx, addrs[0])
	acc.SetCoins(sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(1000000000))))
	ak.SetAccount(suite.ctx, acc)

	expectedSwapIDBytes, err := types.CalculateSwapID(rnh2, binanceAddrs[0], "")
	suite.NoError(err)
	expectedSwapID := types.BytesToHexEncodedString(expectedSwapIDBytes)

	swapID, err := suite.keeper.AddHTLT(suite.ctx, binanceAddrs[0], kavaAddrs[0], "", "", rnh2, timestamp2, coinsSingle, "bnb50000", 80000, false)
	suite.NoError(err)
	suite.Equal(swapID, expectedSwapID)
}

func (suite *HtltTestSuite) TestIterateHtlts() {
	htlts := htlts()
	for _, h := range htlts {
		swapID, err := types.CalculateSwapID(h.RandomNumberHash, h.From, h.SenderOtherChain)
		suite.NoError(err)
		suite.keeper.SetHTLT(suite.ctx, h, swapID)
	}
	res := suite.keeper.GetAllHtlts(suite.ctx)
	suite.Equal(4, len(res))
}

func TestHtltTestSuite(t *testing.T) {
	suite.Run(t, new(HtltTestSuite))
}

func htlts() types.HTLTs {
	var htlts types.HTLTs
	h1 := types.NewHTLT(binanceAddrs[0], kavaAddrs[0], "", "", rnh1, timestamp1, coinsSingle, "bnb50000", 50500, false)
	h2 := types.NewHTLT(binanceAddrs[1], kavaAddrs[1], "", "", rnh2, timestamp2, coinsSingle, "bnb50000", 61500, false)
	h3 := types.NewHTLT(binanceAddrs[2], kavaAddrs[2], "", "", rnh3, timestamp3, coinsSingle, "bnb50000", 72500, false)
	h4 := types.NewHTLT(binanceAddrs[3], kavaAddrs[3], "", "", rnh4, timestamp4, coinsSingle, "bnb50000", 83500, false)
	htlts = append(htlts, h1, h2, h3, h4)
	return htlts
}
