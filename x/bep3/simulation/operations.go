package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	tmtime "github.com/tendermint/tendermint/types/time"

	// "github.com/cosmos/cosmos-sdk/simapp/helpers"

	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateAtomicSwap = "op_weight_msg_create_atomic_swap"
	// OpWeightMsgClaimAtomicSwap  = "op_weight_msg_claim_atomic_swap"
	// OpWeightMsgRefundAtomicSwap = "op_weight_msg_refund_atomic_swap"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, // TODO: ak types.AccountKeeper
	k keeper.Keeper) simulation.WeightedOperations {

	var (
		weightMsgCreateAtomicSwap int
		// weightMsgClaimAtomicSwap  int
		// weightMsgRefundAtomicSwap int
	)

	op1 := SimulateMsgCreateAtomicSwap(k)

	return simulation.WeightedOperations{
		simulation.WeightedOperation{
			Weight: weightMsgCreateAtomicSwap,
			Op:     op1,
		},
		// simulation.NewWeightedOperation(
		// 	weightMsgClaimAtomicSwap,
		// 	SimulateMsgClaimValidator(ak, k),
		// ),
		// simulation.NewWeightedOperation(
		// 	weightMsgRefundAtomicSwap,
		// 	SimulateMsgRefundValidator(ak, k),
		// ),
	}
}

// TODO: Only active assets will be able to be sent
// TODO: Only accounts funded in this denom will be able to send assets
// TODO: Only incoming swaps will be able to be created (sender == deputy)
// TODO: Need to store random numbers otherwise all claims will fail

// SimulateMsgCreateAtomicSwap generates a MsgCreateAtomicSwap with random values
func SimulateMsgCreateAtomicSwap(k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		sender := k.GetBnbDeputyAddress(ctx)
		recipient := simulation.RandomAcc(r, accs).Address

		recipientOtherChain := simulation.RandStringOfLength(r, 43)
		senderOtherChain := simulation.RandStringOfLength(r, 43)

		// Generate cryptographically strong pseudo-random number
		randomNumber, err := types.GenerateSecureRandomNumber()
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		timestamp := tmtime.Now().Unix()
		randomNumberHash := types.CalculateRandomHash(randomNumber.Bytes(), timestamp)

		// Randomly select an asset from supported assets
		assets, found := k.GetAssets(ctx)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}
		asset := assets[r.Intn(len(assets))]
		denom := asset.Denom

		// Check that the sender has coins of this type
		availableAmount := ak.GetAccount(ctx, sender).GetCoins().AmountOf(denom)
		if !availableAmount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// Get a random amount of the available coins
		amount, err := simulation.RandPositiveInt(r, availableAmount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		coin := sdk.NewCoin(denom, amount)
		coins := sdk.NewCoins(coin)
		expectedIncome := coin.String()

		// TODO: Should this be an accepted height span, or one that's between AbsoluteMin/AbsoluteMax?
		minLock := int(k.GetMinBlockLock(ctx))
		maxLock := int(k.GetMaxBlockLock(ctx))
		heightSpan := int64(r.Intn(maxLock-minLock) + minLock)
		crossChain := true

		msg := types.NewMsgCreateAtomicSwap(
			sender, recipient, recipientOtherChain, senderOtherChain, randomNumberHash,
			timestamp, coins, expectedIncome, heightSpan, crossChain)

		// tx := helpers.GenTx(
		// 	[]sdk.Msg{msg},
		// 	fees,
		// 	helpers.DefaultGenTxGas,
		// 	chainID,
		// 	[]uint64{account.GetAccountNumber()},
		// 	[]uint64{account.GetSequence()},
		// 	simAccount.PrivKey,
		// )

		// TODO: app.Deliver(tx) is different
		txResult := app.Deliver(tx)
		// if err != nil {
		// 	return simulation.NoOpMsg(types.ModuleName), nil, err
		// }

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}
