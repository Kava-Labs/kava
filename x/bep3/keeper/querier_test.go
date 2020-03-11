package keeper_test

import (
	"encoding/hex"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	custom = "custom"
)

type QuerierTestSuite struct {
	suite.Suite
	keeper   keeper.Keeper
	app      app.TestApp
	ctx      sdk.Context
	querier  sdk.Querier
	addrs    []sdk.AccAddress
	swapIDs  []cmn.HexBytes
	isSwapID map[string]bool
}

func (suite *QuerierTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})

	// Set up auth GenesisState
	_, addrs := app.GeneratePrivKeyAddressPairs(10)
	coins := []sdk.Coins{}
	for j := 0; j < 10; j++ {
		coins = append(coins, cs(c("bnb", 10000000000), c("ukava", 10000000000)))
	}
	authGS := app.NewAuthGenState(addrs, coins)

	tApp.InitializeFromGenesisStates(
		authGS,
		NewBep3GenStateMulti(),
	)

	suite.ctx = ctx
	suite.app = tApp
	suite.keeper = tApp.GetBep3Keeper()
	suite.querier = keeper.NewQuerier(suite.keeper)
	suite.addrs = addrs

	// Create atomic swaps and save IDs
	var swapIDs []cmn.HexBytes
	isSwapID := make(map[string]bool)
	for i := 0; i < 10; i++ {
		// Set up atomic swap variables
		expireHeight := int64(360)
		amount := cs(c("bnb", int64(50000+i*100)))
		timestamp := ts(0)
		randomNumber, _ := types.GenerateSecureRandomNumber()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

		// Create atomic swap and check err
		err := suite.keeper.CreateAtomicSwap(suite.ctx, randomNumberHash, timestamp, expireHeight,
			suite.addrs[i], suite.addrs[i], binanceAddrs[0].String(), binanceAddrs[1].String(),
			amount, amount.String())
		suite.Nil(err)

		// Calculate swap ID and save
		swapID := types.CalculateSwapID(randomNumberHash, suite.addrs[i], binanceAddrs[0].String())
		swapIDs = append(swapIDs, swapID)
		isSwapID[hex.EncodeToString(swapID)] = true
	}
	suite.swapIDs = swapIDs
	suite.isSwapID = isSwapID
}

func (suite *QuerierTestSuite) TestQueryAtomicSwap() {
	ctx := suite.ctx.WithIsCheckTx(false)

	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAtomicSwap}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryAtomicSwapByID(suite.swapIDs[0])),
	}

	// Execute query and check the []byte result
	bz, err := suite.querier(ctx, []string{types.QueryGetAtomicSwap}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	// Unmarshal the bytes into type atomic swap
	var swap types.AtomicSwap
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &swap))

	// Check the returned atomic swap's ID
	suite.True(suite.isSwapID[hex.EncodeToString(swap.GetSwapID())])
}

func (suite *QuerierTestSuite) TestQueryAtomicSwaps() {
	ctx := suite.ctx.WithIsCheckTx(false)
	// Set up request query
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryGetAtomicSwaps}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(types.NewQueryAtomicSwaps(1, 100)),
	}

	bz, err := suite.querier(ctx, []string{types.QueryGetAtomicSwaps}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var swaps types.AtomicSwaps
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &swaps))

	suite.Equal(len(suite.swapIDs), len(swaps))
	for _, swap := range swaps {
		suite.True(suite.isSwapID[hex.EncodeToString(swap.GetSwapID())])
	}
}

func (suite *QuerierTestSuite) TestQueryParams() {
	ctx := suite.ctx.WithIsCheckTx(false)
	bz, err := suite.querier(ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.Nil(err)
	suite.NotNil(bz)

	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))

	bep3GenesisState := NewBep3GenStateMulti()
	gs := types.GenesisState{}
	types.ModuleCdc.UnmarshalJSON(bep3GenesisState["bep3"], &gs)
	suite.Equal(gs.Params, p)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QuerierTestSuite))
}
