package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	appparams "github.com/kava-labs/kava/app/params"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
)

// Simulation operation weights constants
const (
	OpWeightMsgClaimReward = "op_weight_msg_claim_reward"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak auth.AccountKeeper, sk types.SupplyKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgClaimReward int

	appParams.GetOrGenerate(cdc, OpWeightMsgClaimReward, &weightMsgClaimReward, nil,
		func(_ *rand.Rand) {
			weightMsgClaimReward = appparams.DefaultWeightMsgClaimReward
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgClaimReward,
			SimulateMsgClaimReward(ak, sk, k),
		),
	}
}

// SimulateMsgClaimReward generates a MsgClaimReward
func SimulateMsgClaimReward(ak auth.AccountKeeper, sk types.SupplyKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		// Load only account types that can claim rewards
		validAccounts := make(map[string]bool)
		for _, acc := range accs {
			account := ak.GetAccount(ctx, acc.Address)
			switch account.(type) {
			case *vesting.PeriodicVestingAccount, *auth.BaseAccount: // Valid: BaseAccount, PeriodicVestingAccount
				validAccounts[account.GetAddress().String()] = true
			default: // Invalid: ValidatorVestingAccount, DelayedVestingAccount, ContinuousVestingAccount
				break
			}
		}

		// Load open claims and shuffle them to randomize
		openClaims := types.Claims{}
		k.IterateClaims(ctx, func(claim types.Claim) bool {
			openClaims = append(openClaims, claim)
			return false
		})
		r.Shuffle(len(openClaims), func(i, j int) {
			openClaims[i], openClaims[j] = openClaims[j], openClaims[i]
		})

		kavadistMacc := sk.GetModuleAccount(ctx, kavadist.KavaDistMacc)
		kavadistBalance := kavadistMacc.SpendableCoins(ctx.BlockTime())

		// Find address that has a claim of the same reward denom, then confirm it's distributable
		claimer, claim, found := findValidAccountClaimPair(accs, openClaims, func(acc simulation.Account, claim types.Claim) bool {
			if validAccounts[acc.Address.String()] { // Address must be valid type
				if claim.Owner.Equals(acc.Address) { // Account must be claim owner
					allClaims, found := k.GetActiveClaimsByAddressAndCollateralType(ctx, claim.Owner, claim.CollateralType)
					if found { // found should always be true
						var rewards sdk.Coins
						for _, individualClaim := range allClaims {
							rewards = rewards.Add(individualClaim.Reward)
						}
						if rewards.AmountOf(claim.Reward.Denom).IsPositive() { // Can't distribute 0 coins
							// Validate that kavadist module has enough coins to distribute rewards
							if kavadistBalance.AmountOf(claim.Reward.Denom).GTE(rewards.AmountOf(claim.Reward.Denom)) {
								return true
							}
						}
					}
				}
			}
			return false
		})
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName,
				"no-operation (no accounts currently have fulfillable claims)", "", false, nil), nil, nil
		}

		claimerAcc := ak.GetAccount(ctx, claimer.Address)
		if claimerAcc == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("couldn't find account %s", claimer.Address)
		}

		msg := types.NewMsgClaimReward(claimer.Address, claim.CollateralType)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			sdk.NewCoins(),
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{claimerAcc.GetAccountNumber()},
			[]uint64{claimerAcc.GetSequence()},
			claimer.PrivKey,
		)

		_, result, err := app.Deliver(tx)
		if err != nil {
			// to aid debugging, add the stack trace to the comment field of the returned opMsg
			return simulation.NewOperationMsg(msg, false, fmt.Sprintf("%+v", err)), nil, err
		}

		// to aid debugging, add the result log to the comment field
		return simulation.NewOperationMsg(msg, true, result.Log), nil, nil
	}
}

// findValidAccountClaimPair finds an account and reward claim for which the callback func returns true
func findValidAccountClaimPair(accounts []simulation.Account, claims types.Claims,
	cb func(simulation.Account, types.Claim) bool) (simulation.Account, types.Claim, bool) {
	for _, claim := range claims {
		for _, acc := range accounts {
			if isValid := cb(acc, claim); isValid {
				return acc, claim, true
			}
		}
	}
	return simulation.Account{}, types.Claim{}, false
}
