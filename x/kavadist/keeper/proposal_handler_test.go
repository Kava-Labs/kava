package keeper_test

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

// 	"github.com/Kava-Labs/draco-test/testutil/simapp"
// 	"github.com/Kava-Labs/draco-test/x/kavadist/keeper"
// 	"github.com/Kava-Labs/draco-test/x/kavadist/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// )

// func TestHandleCommunityPoolMultiSpendProposal(t *testing.T) {
// 	app := simapp.New()
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
// 	addr := simapp.AddTestAddrs(app, ctx, 2, sdk.ZeroInt())

// 	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000))
// 	require.NoError(t, simapp.FundAccount(app.BankKeeper, ctx, addr[0], amount))
// 	initPool := app.DistrKeeper.GetFeePool(ctx)
// 	assert.Empty(t, initPool.CommunityPool)

// 	err := app.DistrKeeper.FundCommunityPool(ctx, amount, addr[0])
// 	assert.Nil(t, err)

// 	assert.Equal(t, initPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), app.DistrKeeper.GetFeePool(ctx).CommunityPool)
// 	assert.Empty(t, app.BankKeeper.GetAllBalances(ctx, addr[0]))

// 	proposal := types.NewCommunityPoolMultiSpendProposal("test title", "description", []types.MultiSpendRecipient{
// 		{
// 			Address: addr[0].String(),
// 			Amount:  sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(1000))),
// 		},
// 	})
// 	err = keeper.HandleCommunityPoolMultiSpendProposal(ctx, app.KavadistKeeper, &proposal)
// 	require.Nil(t, err)

// 	coins := app.BankKeeper.GetBalance(ctx, addr[0], "stake")
// 	require.Equal(t, coins.Amount, sdk.NewInt(1000))
// }
