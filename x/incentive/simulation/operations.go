package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
)

var (
	noOpMsg = simulation.NoOpMsg(types.ModuleName)
)

// SimulateMsgClaimReward generates a MsgClaimReward
func SimulateMsgClaimReward(ak auth.AccountKeeper, sk supply.Keeper, k keeper.Keeper) simulation.Operation {
	// handler := incentive.NewHandler(k)

	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {

		// Load only account types that can claim rewards
		var accounts []authexported.Account
		for _, acc := range accs {
			account := ak.GetAccount(ctx, acc.Address)
			switch account.(type) {
			case *vesting.PeriodicVestingAccount, *auth.BaseAccount: // Valid: BaseAccount, PeriodicVestingAccount
				accounts = append(accounts, account)
				break
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

		// Load kavadist module account's current balance
		kavadistMacc := sk.GetModuleAccount(ctx, kavadist.KavaDistMacc)
		kavadistBalance := kavadistMacc.SpendableCoins(ctx.BlockTime())

		// Find address that has a claim of the same reward denom, then confirm it's distributable
		claimer, claim, found := findValidAccountClaimPair(accounts, openClaims, func(acc authexported.Account, claim types.Claim) bool {
			if claim.Owner.Equals(acc.GetAddress()) { // Account must be claim owner
				if claim.Reward.Amount.IsPositive() { // Can't distribute 0 coins
					// Validate that kavadist module has enough coins to distribute the claim
					if kavadistBalance.AmountOf(claim.Reward.Denom).GTE(claim.Reward.Amount) {
						fmt.Println("claim reward:", claim.Reward)
						fmt.Println("kavadist balance:", kavadistBalance.AmountOf(claim.Reward.Denom), claim.Reward.Denom)
						return true
					}
				}
			}
			return false
		})
		if !found {
			return simulation.NewOperationMsgBasic(types.ModuleName,
				"no-operation (no accounts currently have fulfillable claims)", "", false, nil), nil, nil
		}

		msg := types.NewMsgClaimReward(claimer.GetAddress(), claim.Denom)
		if err := msg.ValidateBasic(); err != nil {
			return noOpMsg, nil, fmt.Errorf("expected MsgClaimReward to pass ValidateBasic: %s", err)
		}

		ok := submitMsg(ctx, handler, msg)
		return simulation.NewOperationMsg(msg, ok, ""), nil, nil
	}
}

func submitMsg(ctx sdk.Context, handler sdk.Handler, msg sdk.Msg) (ok bool) {
	ctx, write := ctx.CacheContext()
	result := handler(ctx, msg)
	ok = result.IsOK()
	if ok {
		write()
	} else {
		fmt.Println("Failed:", result.Log)
	}
	return ok
}

// findValidAccountClaimPair finds an account and reward claim for which the callback func returns true
func findValidAccountClaimPair(accounts []authexported.Account, claims types.Claims,
	cb func(authexported.Account, types.Claim) bool) (authexported.Account, types.Claim, bool) {
	for _, claim := range claims {
		for _, acc := range accounts {
			if isValid := cb(acc, claim); isValid {
				return acc, claim, true
			}
		}
	}
	return nil, types.Claim{}, false
}
