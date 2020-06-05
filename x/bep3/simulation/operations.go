package simulation

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/bep3/keeper"
	"github.com/kava-labs/kava/x/bep3/types"
)

var (
	noOpMsg = simtypes.NoOpMsg(types.ModuleName, "", "")
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateAtomicSwap = "op_weight_msg_create_atomic_swap"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc *codec.Codec, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
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
			SimulateMsgCreateAtomicSwap(ak, bk, k),
		),
	}
}

// SimulateMsgCreateAtomicSwap generates a MsgCreateAtomicSwap with random values
func SimulateMsgCreateAtomicSwap(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// Set up deputy address as it's required for all atomic swaps
		deputyAddr := k.GetBnbDeputyAddress(ctx)
		deputyAcc, foundDeputy := simtypes.FindAccount(accs, deputyAddr)
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
		senderOut, asset, found := findValidAccountAssetSupplyPair(accs, supplies, func(acc simtypes.Account, asset types.AssetSupply) bool {
			if asset.CurrentSupply.Amount.IsPositive() {
				if bk.SpendableCoins(ctx, acc.Address).AmountOf(asset.Denom).GT(sdk.NewIntFromUint64(bnbDeputyFixedFee)) {
					return true
				}
			}
			return false
		})

		// Set sender, recipient, and denom depending on swap direction
		var sender simtypes.Account
		var recipient simtypes.Account
		var denom string
		// If an outgoing swap can be created, it's chosen 50% of the time.
		if found && r.Intn(100) < 50 {
			sender = senderOut
			recipient = deputyAcc
			denom = asset.Denom
		} else {
			sender = deputyAcc
			recipient, _ = simtypes.RandomAcc(r, accs)
			// Randomly select an asset from supported assets
			assets, foundAsset := k.GetAssets(ctx)
			if !foundAsset {
				return noOpMsg, nil, fmt.Errorf("no supported assets found")
			}
			denom = assets[r.Intn(len(assets))].Denom
		}

		recipientOtherChain := simtypes.RandStringOfLength(r, 43)
		senderOtherChain := simtypes.RandStringOfLength(r, 43)

		// Generate cryptographically strong pseudo-random number
		randomNumber, err := simtypes.RandPositiveInt(r, sdk.NewInt(math.MaxInt64))
		if err != nil {
			return noOpMsg, nil, err
		}
		// Must use current blocktime instead of 'now' since initial blocktime was randomly generated
		timestamp := ctx.BlockTime().Unix()
		randomNumberHash := types.CalculateRandomHash(randomNumber.BigInt().Bytes(), timestamp)

		// Check that the sender has coins for fee
		fees, err := simtypes.RandomFees(r, ctx, bk.SpendableCoins(ctx, sender.Address))
		if err != nil {
			return noOpMsg, nil, err
		}

		// Get maximum valid amount
		maximumAmount := bk.SpendableCoins(ctx, sender.Address).Sub(fees).AmountOf(denom)
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

		// Get an amount of coins between 0.1 and 2% of total coins
		amount := maximumAmount.Quo(sdk.NewInt(int64(simtypes.RandIntBetween(r, 50, 1000))))
		if amount.LT(sdk.NewIntFromUint64(bnbDeputyFixedFee)) {
			return simtypes.NewOperationMsgBasic(types.ModuleName, fmt.Sprintf("no-operation (all funds exhausted for asset %s)", denom), "", false, nil), nil, nil
		}
		coins := sdk.NewCoins(sdk.NewCoin(denom, amount))

		// We're assuming that sims are run with -NumBlocks=100
		heightSpan := uint64(55)

		msg := types.NewMsgCreateAtomicSwap(
			sender.Address, recipient.Address, recipientOtherChain, senderOtherChain,
			randomNumberHash, timestamp, coins, heightSpan,
		)

		senderAcc := ak.GetAccount(ctx, sender.Address)

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
			return noOpMsg, nil, err
		}

		// Construct a MsgClaimAtomicSwap or MsgRefundAtomicSwap future operation
		var futureOp simtypes.FutureOperation
		swapID := types.CalculateSwapID(msg.RandomNumberHash, msg.From, msg.SenderOtherChain)
		if r.Intn(100) < 50 {
			// Claim future operation
			executionBlock := uint64(ctx.BlockHeight()) + msg.HeightSpan/2
			futureOp = simtypes.FutureOperation{
				BlockHeight: int(executionBlock),
				Op:          operationClaimAtomicSwap(ak, bk, k, swapID, randomNumber.BigInt().Bytes()),
			}
		} else {
			// Refund future operation
			executionBlock := uint64(ctx.BlockHeight()) + msg.HeightSpan
			futureOp = simtypes.FutureOperation{
				BlockHeight: int(executionBlock),
				Op:          operationRefundAtomicSwap(ak, bk, k, swapID),
			}
		}

		return simtypes.NewOperationMsg(msg, true, result.Log), []simtypes.FutureOperation{futureOp}, nil
	}
}

func operationClaimAtomicSwap(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, swapID []byte, randomNumber []byte) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		msg := types.NewMsgClaimAtomicSwap(acc.GetAddress(), swapID, randomNumber)

		fees, err := simtypes.RandomFees(r, ctx, bk.SpendableCoins(ctx, acc.GetAddress()))
		if err != nil {
			return noOpMsg, nil, err
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
			return noOpMsg, nil, err
		}

		return simtypes.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

func operationRefundAtomicSwap(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, swapID []byte) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		acc := ak.GetAccount(ctx, simAccount.Address)

		msg := types.NewMsgRefundAtomicSwap(acc.GetAddress(), swapID)

		fees, err := simtypes.RandomFees(r, ctx, bk.SpendableCoins(ctx, acc.GetAddress()))
		if err != nil {
			return noOpMsg, nil, err
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
			return noOpMsg, nil, err
		}

		return simtypes.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// findValidAccountAssetSupplyPair finds an account for which the callback func returns true
func findValidAccountAssetSupplyPair(accounts []simtypes.Account, supplies types.AssetSupplies,
	cb func(simtypes.Account, types.AssetSupply) bool) (simtypes.Account, types.AssetSupply, bool) {
	for _, supply := range supplies {
		for _, acc := range accounts {
			if isValid := cb(acc, supply); isValid {
				return acc, supply, true
			}
		}
	}
	return simtypes.Account{}, types.AssetSupply{}, false
}
