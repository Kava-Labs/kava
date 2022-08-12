package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

// TransferDelegation moves delegation shares between addresses, while keeping the same validator.
//
// Since the validator is the same, underlying staking tokens are not transferred between the bonded and not bonded pools.
// Vesting periods for delegated tokens will not be transferred to the new delegator.
// The sending delegation must not have any active redelegations.
// A validator cannot reduce self delegated shares below its min self delegation.
func (k Keeper) TransferDelegation(ctx sdk.Context, valAddr sdk.ValAddress, fromDelegator, toDelegator sdk.AccAddress, shares sdk.Dec) error {
	// Redelegations link a delegation to it's previous validator so slashes are propagated to the new validator.
	// If the delegation is transferred to a new owner, the redelegation object must be updated.
	// For expediency all transfers with redelegations are blocked.
	if invalid := k.stakingKeeper.HasReceivingRedelegation(ctx, fromDelegator, valAddr); invalid {
		return types.ErrRedelegationsNotCompleted
	}

	// ensure validator exists
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.ErrNoValidatorFound
	}

	fromDelegation, found := k.stakingKeeper.GetDelegation(ctx, fromDelegator, valAddr)
	if !found {
		return types.ErrNoDelegatorForAddress
	}

	if shares.IsNil() || shares.LT(sdk.ZeroDec()) {
		return sdkerrors.Wrap(types.ErrInvalidRequest, "cannot transfer nil or negative shares")
	}
	if fromDelegation.Shares.LT(shares) {
		return sdkerrors.Wrapf(types.ErrNotEnoughDelegationShares, "%s < %s", fromDelegation.Shares, shares)
	}

	isValidatorOperator := fromDelegator.Equals(valAddr)
	if isValidatorOperator {
		if isBelowMinSelfDelegation(validator, fromDelegation.Shares.Sub(shares)) {
			return types.ErrSelfDelegationBelowMinimum
		}
	}

	toDelegation, foundToDelegation := k.stakingKeeper.GetDelegation(ctx, toDelegator, valAddr)
	if !foundToDelegation {
		toDelegation = stakingtypes.NewDelegation(toDelegator, valAddr, sdk.ZeroDec())
	}

	k.stakingKeeper.BeforeDelegationSharesModified(ctx, fromDelegator, valAddr)
	if foundToDelegation {
		k.stakingKeeper.BeforeDelegationSharesModified(ctx, toDelegator, valAddr)
	} else {
		k.stakingKeeper.BeforeDelegationCreated(ctx, toDelegator, valAddr)
	}

	fromDelegation.Shares = fromDelegation.Shares.Sub(shares)
	toDelegation.Shares = toDelegation.Shares.Add(shares)

	fromDelegationEmpty := fromDelegation.Shares.IsZero()
	if fromDelegationEmpty {
		k.stakingKeeper.RemoveDelegation(ctx, fromDelegation) // calls BeforeDelegationRemoved hook
	} else {
		k.stakingKeeper.SetDelegation(ctx, fromDelegation)
	}
	k.stakingKeeper.SetDelegation(ctx, toDelegation)

	// Call 'After' hooks after both delegations have been set, to ensure other modules have a valid staking state to query from.
	if !fromDelegationEmpty {
		k.stakingKeeper.AfterDelegationModified(ctx, fromDelegator, valAddr)
	}
	k.stakingKeeper.AfterDelegationModified(ctx, toDelegator, valAddr)

	return nil
}

func isBelowMinSelfDelegation(validator stakingtypes.ValidatorI, shares sdk.Dec) bool {
	return validator.TokensFromShares(shares).TruncateInt().LT(validator.GetMinSelfDelegation())
}
