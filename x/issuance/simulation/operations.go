package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simappparams "github.com/kava-labs/kava/app/params"

	"github.com/kava-labs/kava/x/issuance/keeper"
	"github.com/kava-labs/kava/x/issuance/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgIssue  = "op_weight_msg_issue"
	OpWeightMsgRedeem = "op_weight_msg_redeem"
	OpWeightMsgBlock  = "op_weight_msg_block"
	OpWeightMsgPause  = "op_weight_msg_pause"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgIssue  int
		weightMsgReedem int
		weightMsgBlock  int
		weightMsgPause  int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgIssue, &weightMsgIssue, nil,
		func(_ *rand.Rand) {
			weightMsgIssue = simappparams.DefaultWeightMsgIssue
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgRedeem, &weightMsgReedem, nil,
		func(_ *rand.Rand) {
			weightMsgReedem = simappparams.DefaultWeightMsgRedeem
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgBlock, &weightMsgBlock, nil,
		func(_ *rand.Rand) {
			weightMsgBlock = simappparams.DefaultWeightMsgBlock
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgPause, &weightMsgPause, nil,
		func(_ *rand.Rand) {
			weightMsgPause = simappparams.DefaultWeightMsgPause
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgIssue,
			SimulateMsgIssueTokens(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgReedem,
			SimulateMsgRedeemTokens(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBlock,
			SimulateMsgBlockAddress(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgPause,
			SimulateMsgPause(ak, k),
		),
	}
}

// SimulateMsgIssueTokens generates a MsgIssueTokens with random values
func SimulateMsgIssueTokens(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// shuffle the assets and get a random one
		assets := k.GetParams(ctx).Assets
		r.Shuffle(len(assets), func(i, j int) {
			assets[i], assets[j] = assets[j], assets[i]
		})
		asset := assets[0]

		if asset.Paused {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// make sure owner account exists
		ownerSimAcc, found := simulation.FindAccount(accs, asset.Owner)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("asset owner not found: %s", asset)
		}

		// issue new tokens to the owner 50% of the time so we have funds to redeem
		ownerAcc := ak.GetAccount(ctx, asset.Owner)
		recipient := ownerAcc
		if r.Intn(2) == 0 {
			simAccount, _ := simulation.RandomAcc(r, accs)
			recipient = ak.GetAccount(ctx, simAccount.Address)
		}
		if recipient == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		randomAmount := simulation.RandIntBetween(r, 10000000, 1000000000000)
		msg := types.NewMsgIssueTokens(asset.Owner, sdk.NewCoin(asset.Denom, sdk.NewInt(int64(randomAmount))), recipient.GetAddress())
		spendableCoins := ownerAcc.SpendableCoins(ctx.BlockTime())
		fees, err := simulation.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{ownerAcc.GetAccountNumber()},
			[]uint64{ownerAcc.GetSequence()},
			ownerSimAcc.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgRedeemTokens generates a MsgRedeemTokens with random values
func SimulateMsgRedeemTokens(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// shuffle the assets and get a random one
		assets := k.GetParams(ctx).Assets
		r.Shuffle(len(assets), func(i, j int) {
			assets[i], assets[j] = assets[j], assets[i]
		})
		asset := assets[0]
		if asset.Paused {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// make sure owner account exists
		ownerSimAcc, found := simulation.FindAccount(accs, asset.Owner)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("asset owner not found: %s", asset)
		}

		ownerAcc := ak.GetAccount(ctx, asset.Owner)

		spendableCoinAmount := ownerAcc.SpendableCoins(ctx.BlockTime()).AmountOf(asset.Denom)
		if spendableCoinAmount.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		var redeemAmount sdk.Int
		if spendableCoinAmount.Equal(sdk.OneInt()) {
			redeemAmount = sdk.OneInt()
		}

		redeemAmount = sdk.NewInt(int64(simulation.RandIntBetween(r, 1, int(spendableCoinAmount.Int64()))))

		msg := types.NewMsgRedeemTokens(asset.Owner, sdk.NewCoin(asset.Denom, redeemAmount))
		spendableCoins := ownerAcc.SpendableCoins(ctx.BlockTime()).Sub(sdk.NewCoins(sdk.NewCoin(asset.Denom, redeemAmount)))
		fees, err := simulation.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{ownerAcc.GetAccountNumber()},
			[]uint64{ownerAcc.GetSequence()},
			ownerSimAcc.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgBlockAddress generates a MsgBlockAddress with random values
func SimulateMsgBlockAddress(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// shuffle the assets and get a random one
		assets := k.GetParams(ctx).Assets
		r.Shuffle(len(assets), func(i, j int) {
			assets[i], assets[j] = assets[j], assets[i]
		})
		asset := assets[0]

		// make sure owner account exists
		ownerSimAcc, found := simulation.FindAccount(accs, asset.Owner)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("asset owner not found: %s", asset)
		}
		ownerAcc := ak.GetAccount(ctx, asset.Owner)

		// find an account to block
		simAccount, _ := simulation.RandomAcc(r, accs)
		blockedAccount := ak.GetAccount(ctx, simAccount.Address)
		if blockedAccount.GetAddress().Equals(asset.Owner) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgBlockAddress(asset.Owner, asset.Denom, blockedAccount.GetAddress())
		spendableCoins := ownerAcc.SpendableCoins(ctx.BlockTime())
		fees, err := simulation.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{ownerAcc.GetAccountNumber()},
			[]uint64{ownerAcc.GetSequence()},
			ownerSimAcc.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgPause generates a MsgChangePauseStatus with random values
func SimulateMsgPause(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// shuffle the assets and get a random one
		assets := k.GetParams(ctx).Assets
		r.Shuffle(len(assets), func(i, j int) {
			assets[i], assets[j] = assets[j], assets[i]
		})
		asset := assets[0]

		// make sure owner account exists
		ownerSimAcc, found := simulation.FindAccount(accs, asset.Owner)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("asset owner not found: %s", asset)
		}
		ownerAcc := ak.GetAccount(ctx, asset.Owner)

		// set status to paused/un-paused
		status := r.Intn(2) == 0

		msg := types.NewMsgChangePauseStatus(asset.Owner, asset.Denom, status)
		spendableCoins := ownerAcc.SpendableCoins(ctx.BlockTime())
		fees, err := simulation.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{ownerAcc.GetAccountNumber()},
			[]uint64{ownerAcc.GetSequence()},
			ownerSimAcc.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}
