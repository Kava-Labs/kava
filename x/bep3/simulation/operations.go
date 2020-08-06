package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

var (
	noOpMsg      = simulation.NoOpMsg(types.ModuleName)
	randomNumber = []byte{114, 21, 74, 180, 81, 92, 21, 91, 173, 164, 143, 111, 120, 58, 241, 58, 40, 22, 59, 133, 102, 233, 55, 149, 12, 199, 231, 63, 122, 23, 88, 9}
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateAtomicSwap = "op_weight_msg_create_atomic_swap"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightCreateAtomicSwap int

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateAtomicSwap, &weightCreateAtomicSwap, nil,
		func(_ *rand.Rand) {
			weightCreateAtomicSwap = appparams.DefaultWeightMsgCreateAtomicSwap
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightCreateAtomicSwap,
			SimulateMsgCreateAtomicSwap(ak, k),
		),
	}
}

// SimulateMsgCreateAtomicSwap generates a MsgCreateAtomicSwap with random values
func SimulateMsgCreateAtomicSwap(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		// Get asset supplies and shuffle them
		assets, found := k.GetAssets(ctx)
		if !found {
			return noOpMsg, nil, nil
		}
		r.Shuffle(len(assets), func(i, j int) {
			assets[i], assets[j] = assets[j], assets[i]
		})
		senderOutgoing, selectedAsset, found := findValidAccountAssetPair(accs, assets, func(simAcc simulation.Account, asset types.AssetParam) bool {
			supply, found := k.GetAssetSupply(ctx, asset.Denom)
			if !found {
				return false
			}
			if supply.CurrentSupply.Amount.IsPositive() {
				authAcc := ak.GetAccount(ctx, simAcc.Address)
				// deputy cannot be sender of outgoing swap
				if authAcc.GetAddress().Equals(asset.DeputyAddress) {
					return false
				}
				// Search for an account that holds coins received by an atomic swap
				minAmountPlusFee := asset.MinSwapAmount.Add(asset.FixedFee)
				if authAcc.SpendableCoins(ctx.BlockTime()).AmountOf(asset.Denom).GT(minAmountPlusFee) {
					return true
				}
			}
			return false
		})
		var sender simulation.Account
		var recipient simulation.Account
		var asset types.AssetParam

		// If an outgoing swap can be created, it's chosen 50% of the time.
		if found && r.Intn(100) < 50 {
			deputy, found := simulation.FindAccount(accs, selectedAsset.DeputyAddress)
			if !found {
				return noOpMsg, nil, nil
			}
			sender = senderOutgoing
			recipient = deputy
			asset = selectedAsset
		} else {
			// if an outgoing swap cannot be created or was not selected, simulate an incoming swap
			assets, _ := k.GetAssets(ctx)
			asset = assets[r.Intn(len(assets))]
			var eligibleAccs []simulation.Account
			for _, simAcc := range accs {
				// don't allow recipient of incoming swap to be the deputy
				if simAcc.Address.Equals(asset.DeputyAddress) {
					continue
				}
				eligibleAccs = append(eligibleAccs, simAcc)
			}
			recipient, _ = simulation.RandomAcc(r, eligibleAccs)
			deputy, found := simulation.FindAccount(accs, asset.DeputyAddress)
			if !found {
				return noOpMsg, nil, nil
			}
			sender = deputy

		}

		recipientOtherChain := simulation.RandStringOfLength(r, 43)
		senderOtherChain := simulation.RandStringOfLength(r, 43)

		// Use same random number for determinism
		timestamp := ctx.BlockTime().Unix()
		randomNumberHash := types.CalculateRandomHash(randomNumber, timestamp)

		// Check that the sender has coins for fee
		senderAcc := ak.GetAccount(ctx, sender.Address)
		fees, err := simulation.RandomFees(r, ctx, senderAcc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// Get maximum valid amount
		maximumAmount := senderAcc.SpendableCoins(ctx.BlockTime()).Sub(fees).AmountOf(asset.Denom)
		// The maximum amount for outgoing swaps is limited by the asset's current supply
		if recipient.Address.Equals(asset.DeputyAddress) {
			assetSupply, foundAssetSupply := k.GetAssetSupply(ctx, asset.Denom)
			if !foundAssetSupply {
				return noOpMsg, nil, fmt.Errorf("no asset supply found for %s", asset.Denom)
			}
			if maximumAmount.GT(assetSupply.CurrentSupply.Amount.Sub(assetSupply.OutgoingSupply.Amount)) {
				maximumAmount = assetSupply.CurrentSupply.Amount.Sub(assetSupply.OutgoingSupply.Amount)
			}
		}

		// The maximum amount for all swaps is limited by the total max limit
		if maximumAmount.GT(asset.MaxSwapAmount) {
			maximumAmount = asset.MaxSwapAmount
		}

		// Get an amount of coins between 0.1 and 2% of total coins
		amount := maximumAmount.Quo(sdk.NewInt(int64(simulation.RandIntBetween(r, 50, 1000))))
		minAmountPlusFee := asset.MinSwapAmount.Add(asset.FixedFee)
		if amount.LT(minAmountPlusFee) {
			return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (all funds exhausted for asset %s)", asset.Denom), "", false, nil), nil, nil
		}
		coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, amount))

		// We're assuming that sims are run with -NumBlocks=100
		heightSpan := uint64(simulation.RandIntBetween(r, int(asset.MinBlockLock), int(asset.MaxBlockLock)))

		msg := types.NewMsgCreateAtomicSwap(
			sender.Address, recipient.Address, recipientOtherChain, senderOtherChain,
			randomNumberHash, timestamp, coins, heightSpan,
		)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{senderAcc.GetAccountNumber()},
			[]uint64{senderAcc.GetSequence()},
			sender.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// Construct a MsgClaimAtomicSwap or MsgRefundAtomicSwap future operation
		var futureOp simulation.FutureOperation
		swapID := types.CalculateSwapID(msg.RandomNumberHash, msg.From, msg.SenderOtherChain)
		if r.Intn(100) < 50 {
			// Claim future operation - choose between next block and the block before height span
			executionBlock := uint64(
				int(ctx.BlockHeight()+1) + r.Intn(int(heightSpan-1)))

			futureOp = simulation.FutureOperation{
				BlockHeight: int(executionBlock),
				Op:          operationClaimAtomicSwap(ak, k, swapID, randomNumber),
			}
		} else {
			// Refund future operation
			executionBlock := uint64(ctx.BlockHeight()) + msg.HeightSpan
			futureOp = simulation.FutureOperation{
				BlockHeight: int(executionBlock),
				Op:          operationRefundAtomicSwap(ak, k, swapID),
			}
		}

		return simulation.NewOperationMsg(msg, true, result.Log), []simulation.FutureOperation{futureOp}, nil
	}
}

func operationClaimAtomicSwap(ak types.AccountKeeper, k keeper.Keeper, swapID []byte, randomNumber []byte) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		swap, found := k.GetAtomicSwap(ctx, swapID)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("cannot claim: swap with ID %s not found", swapID)
		}
		// check that asset supply supports claiming (it could have changed due to a param change proposal)
		// use CacheContext so changes don't take effect
		cacheCtx, _ := ctx.CacheContext()
		switch swap.Direction {
		case types.Incoming:
			err := k.DecrementIncomingAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to decrement incoming asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
			err = k.IncrementCurrentAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to increment current asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		case types.Outgoing:
			err := k.DecrementOutgoingAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to decrement outgoing asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
			err = k.DecrementCurrentAssetSupply(cacheCtx, swap.Amount[0])
			if err != nil {
				return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - unable to decrement current asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		}

		asset, err := k.GetAsset(ctx, swap.Amount[0].Denom)
		if err != nil {
			return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - asset not found %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
		}
		supply, found := k.GetAssetSupply(ctx, asset.Denom)
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not claim - asset supply not found %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
		}
		if asset.SupplyLimit.Limit.LT(supply.CurrentSupply.Amount.Add(swap.Amount[0].Amount)) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgClaimAtomicSwap(acc.GetAddress(), swapID, randomNumber)
		fees, err := simulation.RandomFees(r, ctx, acc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAccount.PrivKey,
		)
		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

func operationRefundAtomicSwap(ak types.AccountKeeper, k keeper.Keeper, swapID []byte) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		swap, found := k.GetAtomicSwap(ctx, swapID)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("cannot refund: swap with ID %s not found", swapID)
		}
		cacheCtx, _ := ctx.CacheContext()
		switch swap.Direction {
		case types.Incoming:
			if err := k.DecrementIncomingAssetSupply(cacheCtx, swap.Amount[0]); err != nil {
				return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not refund - unable to decrement incoming asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		case types.Outgoing:
			if err := k.DecrementOutgoingAssetSupply(cacheCtx, swap.Amount[0]); err != nil {
				return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (could not refund - unable to decrement outgoing asset supply %s)", swap.Amount[0].Denom), "", false, nil), nil, nil
			}
		}

		msg := types.NewMsgRefundAtomicSwap(acc.GetAddress(), swapID)

		fees, err := simulation.RandomFees(r, ctx, acc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{acc.GetAccountNumber()},
			[]uint64{acc.GetSequence()},
			simAccount.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// findValidAccountAssetSupplyPair finds an account for which the callback func returns true
func findValidAccountAssetPair(accounts []simulation.Account, assets types.AssetParams,
	cb func(simulation.Account, types.AssetParam) bool) (simulation.Account, types.AssetParam, bool) {
	for _, asset := range assets {
		for _, acc := range accounts {
			if isValid := cb(acc, asset); isValid {
				return acc, asset, true
			}
		}
	}
	return simulation.Account{}, types.AssetParam{}, false
}
