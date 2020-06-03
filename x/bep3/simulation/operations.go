package simulation

import (
	"fmt"
	"math"
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
	noOpMsg = simulation.NoOpMsg(types.ModuleName)
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

		// Set up deputy address as it's required for all atomic swaps
		deputyAddr := k.GetBnbDeputyAddress(ctx)
		deputyAcc, foundDeputy := simulation.FindAccount(accs, deputyAddr)
		if !foundDeputy {
			return noOpMsg, nil, nil
		}

		// Get asset supplies and shuffle them
		supplies := k.GetAllAssetSupplies(ctx)
		r.Shuffle(len(supplies), func(i, j int) {
			supplies[i], supplies[j] = supplies[j], supplies[i]
		})

		// Search for an account that holds coins received by an atomic swap
		bnbDeputyFixedFee := k.GetBnbDeputyFixedFee(ctx)
		minAmount := k.GetMinAmount(ctx)
		minAmountPlusFee := sdk.NewIntFromUint64(minAmount + bnbDeputyFixedFee)
		senderOut, asset, found := findValidAccountAssetSupplyPair(accs, supplies, func(acc simulation.Account, asset types.AssetSupply) bool {
			if asset.CurrentSupply.Amount.IsPositive() {
				authAcc := ak.GetAccount(ctx, acc.Address)
				if authAcc.SpendableCoins(ctx.BlockTime()).AmountOf(asset.Denom).GT(minAmountPlusFee) {
					return true
				}
			}
			return false
		})

		// Set sender, recipient, and denom depending on swap direction
		var sender simulation.Account
		var recipient simulation.Account
		var denom string
		// If an outgoing swap can be created, it's chosen 50% of the time.
		if found && r.Intn(100) < 50 {
			sender = senderOut
			recipient = deputyAcc
			denom = asset.Denom
		} else {
			sender = deputyAcc
			recipient, _ = simulation.RandomAcc(r, accs)
			// Randomly select an asset from supported assets
			assets, foundAsset := k.GetAssets(ctx)
			if !foundAsset {
				return noOpMsg, nil, fmt.Errorf("no supported assets found")
			}
			denom = assets[r.Intn(len(assets))].Denom
		}

		recipientOtherChain := simulation.RandStringOfLength(r, 43)
		senderOtherChain := simulation.RandStringOfLength(r, 43)

		// Generate cryptographically strong pseudo-random number
		randomNumber, err := simulation.RandPositiveInt(r, sdk.NewInt(math.MaxInt64))
		if err != nil {
			return noOpMsg, nil, err
		}
		// Must use current blocktime instead of 'now' since initial blocktime was randomly generated
		timestamp := ctx.BlockTime().Unix()
		randomNumberHash := types.CalculateRandomHash(randomNumber.BigInt().Bytes(), timestamp)

		// Check that the sender has coins for fee
		senderAcc := ak.GetAccount(ctx, sender.Address)
		fees, err := simulation.RandomFees(r, ctx, senderAcc.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		// Get maximum valid amount
		maximumAmount := senderAcc.SpendableCoins(ctx.BlockTime()).Sub(fees).AmountOf(denom)
		// The maximum amount for outgoing swaps is limited by the asset's current supply
		if recipient.Equals(deputyAcc) {
			assetSupply, foundAssetSupply := k.GetAssetSupply(ctx, []byte(denom))
			if !foundAssetSupply {
				return noOpMsg, nil, fmt.Errorf("no asset supply found")
			}
			if maximumAmount.GT(assetSupply.CurrentSupply.Amount) {
				maximumAmount = assetSupply.CurrentSupply.Amount
			}
		}

		// The maximum amount for all swaps is limited by the total max limit
		maxAmountLimit := sdk.NewIntFromUint64(k.GetMaxAmount(ctx))
		if maximumAmount.GT(maxAmountLimit) {
			maximumAmount = maxAmountLimit
		}

		// Get an amount of coins between 0.1 and 2% of total coins
		amount := maximumAmount.Quo(sdk.NewInt(int64(simulation.RandIntBetween(r, 50, 1000))))
		if amount.LT(minAmountPlusFee) {
			return simulation.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (all funds exhausted for asset %s)", denom), "", false, nil), nil, nil
		}
		coins := sdk.NewCoins(sdk.NewCoin(denom, amount))

		// We're assuming that sims are run with -NumBlocks=100
		heightSpan := uint64(55)

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
			// Claim future operation
			executionBlock := uint64(ctx.BlockHeight()) + msg.HeightSpan/2
			futureOp = simulation.FutureOperation{
				BlockHeight: int(executionBlock),
				Op:          operationClaimAtomicSwap(ak, k, swapID, randomNumber.BigInt().Bytes()),
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
func findValidAccountAssetSupplyPair(accounts []simulation.Account, supplies types.AssetSupplies,
	cb func(simulation.Account, types.AssetSupply) bool) (simulation.Account, types.AssetSupply, bool) {
	for _, supply := range supplies {
		for _, acc := range accounts {
			if isValid := cb(acc, supply); isValid {
				return acc, supply, true
			}
		}
	}
	return simulation.Account{}, types.AssetSupply{}, false
}
