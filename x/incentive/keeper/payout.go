package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/kava-labs/kava/x/incentive/types"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
)

// PayoutClaim sends the timelocked claim coins to the input address
func (k Keeper) PayoutClaim(ctx sdk.Context, addr sdk.AccAddress, denom string, id uint64) error {
	claim, found := k.GetClaim(ctx, addr, denom, id)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimNotFound, "id: %d, denom %s, address: %s", id, denom, addr)
	}
	claimPeriod, found := k.GetClaimPeriod(ctx, id, denom)
	if !found {
		return sdkerrors.Wrapf(types.ErrClaimPeriodNotFound, "id: %d, denom: %s", id, denom)
	}
	err := k.SendTimeLockedCoinsToAccount(ctx, types.IncentiveMacc, addr, sdk.NewCoins(claim.Reward), int64(claimPeriod.TimeLock.Seconds()))
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeySender, fmt.Sprintf("%s", addr)),
		),
	)
	return nil
}

// SendTimeLockedCoinsToAccount sends time-locked coins from the input module account to the recipient. If the recipients account is not a vesting account, it is converted to a periodic vesting account and the coins are added to the vesting balance as a vesting period with the input length.
func (k Keeper) SendTimeLockedCoinsToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins, length int64) error {
	macc := k.supplyKeeper.GetModuleAccount(ctx, senderModule)
	if !macc.GetCoins().IsAllGTE(amt) {
		return sdkerrors.Wrapf(types.ErrInsufficientModAccountBalance, "%s", senderModule)
	}

	// 0. Get the account from the account keeper and do a type switch, error if it's a validator vesting account or module account (can make this work for validator vesting later if necessary)
	acc := k.accountKeeper.GetAccount(ctx, recipientAddr)

	switch acc.(type) {
	case *validatorvesting.ValidatorVestingAccount, supplyExported.ModuleAccountI:
		return sdkerrors.Wrapf(types.ErrInvalidAccountType, "%T", acc)
	case *vesting.PeriodicVestingAccount:
		return k.SendTimeLockedCoinsToPeriodicVestingAccount(ctx, senderModule, recipientAddr, amt, length)
	case *auth.BaseAccount:
		return k.SendTimeLockedCoinsToBaseAccount(ctx, senderModule, recipientAddr, amt, length)
	default:
		return sdkerrors.Wrapf(types.ErrInvalidAccountType, "%T", acc)
	}
}

// SendTimeLockedCoinsToPeriodicVestingAccount sends time-locked coins from the input module account to the recipient
func (k Keeper) SendTimeLockedCoinsToPeriodicVestingAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins, length int64) error {
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
	if err != nil {
		return err
	}
	k.addCoinsToVestingSchedule(ctx, recipientAddr, amt, length)
	return nil
}

// SendTimeLockedCoinsToBaseAccount sends time-locked coins from the input module account to the recipient, converting the recipient account to a vesting account
func (k Keeper) SendTimeLockedCoinsToBaseAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins, length int64) error {
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
	if err != nil {
		return err
	}
	acc := k.accountKeeper.GetAccount(ctx, recipientAddr)
	// transition the account to a periodic vesting account:
	bacc := authtypes.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	newPeriods := vesting.Periods{types.NewPeriod(amt, length)}
	bva, err := vesting.NewBaseVestingAccount(bacc, amt, ctx.BlockTime().Unix()+length)
	if err != nil {
		return err
	}
	pva := vesting.NewPeriodicVestingAccountRaw(bva, ctx.BlockTime().Unix(), newPeriods)
	k.accountKeeper.SetAccount(ctx, pva)
	return nil
}

// DeleteExpiredClaimsAndClaimPeriods deletes expired claim periods and their associated claims
func (k Keeper) DeleteExpiredClaimsAndClaimPeriods(ctx sdk.Context) {
	k.IterateClaimPeriods(ctx, func(cp types.ClaimPeriod) (stop bool) {
		if !cp.End.Before(ctx.BlockTime()) {
			return false
		}
		k.IterateClaims(ctx, func(c types.Claim) (stop bool) {
			if !(c.Denom == cp.Denom && c.ClaimPeriodID == cp.ID) {
				return false
			}
			k.DeleteClaim(ctx, c.Owner, c.Denom, c.ClaimPeriodID)
			return false
		})
		k.DeleteClaimPeriod(ctx, cp.ID, cp.Denom)
		return false
	})
}

// GetClaimsByAddressAndDenom returns all claims for a specific user and address and a bool for if any were found
func (k Keeper) GetClaimsByAddressAndDenom(ctx sdk.Context, addr sdk.AccAddress, denom string) (claims types.Claims, found bool) {
	found = false
	k.IterateClaimPeriods(ctx, func(cp types.ClaimPeriod) (stop bool) {
		if cp.Denom != denom {
			return false
		}
		c, hasClaim := k.GetClaim(ctx, addr, cp.Denom, cp.ID)
		if !hasClaim {
			return false
		}
		found = true
		claims = append(claims, c)
		return false
	})
	return claims, found
}

// addCoinsToVestingSchedule adds coins to the input account's vesting schedule where length is the amount of time (from the current block time), in seconds, that the coins will be vesting for
// the input address must be a periodic vesting account
func (k Keeper) addCoinsToVestingSchedule(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins, length int64) {
	acc := k.accountKeeper.GetAccount(ctx, addr)
	vacc := acc.(*vesting.PeriodicVestingAccount)
	// Add the new vesting coins to OriginalVesting
	vacc.OriginalVesting = vacc.OriginalVesting.Add(amt...)
	// update vesting periods
	if vacc.EndTime < ctx.BlockTime().Unix() {
		// edge case one - the vesting account's end time is in the past (ie, all previous vesting periods have completed)
		// append a new period to the vesting account, update the end time, update the account in the store and return
		newPeriodLength := (ctx.BlockTime().Unix() - vacc.EndTime) + length
		newPeriod := types.NewPeriod(amt, newPeriodLength)
		vacc.VestingPeriods = append(vacc.VestingPeriods, newPeriod)
		vacc.EndTime = ctx.BlockTime().Unix() + length
		k.accountKeeper.SetAccount(ctx, vacc)
		return
	}
	if vacc.StartTime > ctx.BlockTime().Unix() {
		// edge case two - the vesting account's start time is in the future (all periods have not started)
		// update the start time to now and adjust the period lengths in place - a new period will be inserted in the next code block
		updatedPeriods := vesting.Periods{}
		for i, period := range vacc.VestingPeriods {
			updatedPeriod := period
			if i == 0 {
				updatedPeriod = types.NewPeriod(period.Amount, (vacc.StartTime-ctx.BlockTime().Unix())+period.Length)
			}
			updatedPeriods = append(updatedPeriods, updatedPeriod)
		}
		vacc.VestingPeriods = updatedPeriods
		vacc.StartTime = ctx.BlockTime().Unix()
	}

	// logic for inserting a new vesting period into the existing vesting schedule
	totalPeriodLength := types.GetTotalVestingPeriodLength(vacc.VestingPeriods)
	proposedEndTime := ctx.BlockTime().Unix() + length
	if totalPeriodLength < length {
		// in the case that the proposed length is longer than the sum of all previous period lengths, create a new period with length equal to the difference between the proposed length and the previous total length
		newPeriodLength := length - totalPeriodLength
		newPeriod := types.NewPeriod(amt, newPeriodLength)
		vacc.VestingPeriods = append(vacc.VestingPeriods, newPeriod)
		// update the end time so that the sum of all period lengths equals endTime - startTime
		vacc.EndTime = proposedEndTime
	} else {
		// In the case that the proposed length is less than or equal to the sum of all previous period lengths, insert the period and update other periods as necessary.
		// EXAMPLE (l is length, a is amount)
		// Original Periods: {[l: 1 a: 1], [l: 2, a: 1], [l:8, a:3], [l: 5, a: 3]}
		// Period we want to insert [l: 5, a: x]
		// Expected result:
		// {[l: 1, a: 1], [l:2, a: 1], [l:2, a:x], [l:6, a:3], [l:5, a:3]}

		newPeriods := vesting.Periods{}
		lengthCounter := int64(0)
		appendRemaining := false
		for _, period := range vacc.VestingPeriods {
			if appendRemaining {
				newPeriods = append(newPeriods, period)
				continue
			}
			lengthCounter += period.Length
			if lengthCounter < length {
				newPeriods = append(newPeriods, period)
			} else if lengthCounter == length {
				newPeriod := types.NewPeriod(period.Amount.Add(amt...), period.Length)
				newPeriods = append(newPeriods, newPeriod)
				appendRemaining = true
			} else {
				newPeriod := types.NewPeriod(amt, length-types.GetTotalVestingPeriodLength(newPeriods))
				previousPeriod := types.NewPeriod(period.Amount, period.Length-newPeriod.Length)
				newPeriods = append(newPeriods, newPeriod, previousPeriod)
				appendRemaining = true
			}
		}
		vacc.VestingPeriods = newPeriods
	}
	k.accountKeeper.SetAccount(ctx, vacc)
	return
}
