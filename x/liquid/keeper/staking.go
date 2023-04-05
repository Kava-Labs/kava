package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/x/liquid/types"
)

// TransferDelegation moves some delegation shares between addresses, while keeping the same validator.
//
// Internally shares are unbonded, tokens moved then bonded again. This limits only vested tokens from being transferred.
// The sending delegation must not have any active redelegations.
// A validator cannot reduce self delegated shares below its min self delegation.
// Attempting to transfer zero shares will error.
func (k Keeper) TransferDelegation(ctx sdk.Context, valAddr sdk.ValAddress, fromDelegator, toDelegator sdk.AccAddress, shares sdk.Dec) (sdk.Dec, error) {
	// Redelegations link a delegation to it's previous validator so slashes are propagated to the new validator.
	// If the delegation is transferred to a new owner, the redelegation object must be updated.
	// For expediency all transfers with redelegations are blocked.
	if k.stakingKeeper.HasReceivingRedelegation(ctx, fromDelegator, valAddr) {
		return sdk.Dec{}, types.ErrRedelegationsNotCompleted
	}

	if shares.IsNil() || shares.LT(sdk.ZeroDec()) {
		return sdk.Dec{}, errorsmod.Wrap(types.ErrUntransferableShares, "nil or negative shares")
	}
	if shares.Equal(sdk.ZeroDec()) {
		// Block 0 transfers to reduce edge cases.
		return sdk.Dec{}, errorsmod.Wrap(types.ErrUntransferableShares, "zero shares")
	}

	fromDelegation, found := k.stakingKeeper.GetDelegation(ctx, fromDelegator, valAddr)
	if !found {
		return sdk.Dec{}, types.ErrNoDelegatorForAddress
	}
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return sdk.Dec{}, types.ErrNoValidatorFound
	}
	// Prevent validators from reducing their self delegation below the min.
	isValidatorOperator := fromDelegator.Equals(valAddr)
	if isValidatorOperator {
		if isBelowMinSelfDelegation(validator, fromDelegation.Shares.Sub(shares)) {
			return sdk.Dec{}, types.ErrSelfDelegationBelowMinimum
		}
	}

	returnAmount, err := k.fastUndelegate(ctx, valAddr, fromDelegator, shares)
	if err != nil {
		return sdk.Dec{}, err
	}
	returnCoins := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), returnAmount))

	if err := k.bankKeeper.SendCoins(ctx, fromDelegator, toDelegator, returnCoins); err != nil {
		return sdk.Dec{}, err
	}
	receivedShares, err := k.delegateFromAccount(ctx, valAddr, toDelegator, returnAmount)
	if err != nil {
		return sdk.Dec{}, err
	}

	return receivedShares, nil
}

// isBelowMinSelfDelegation check if the supplied shares, converted to tokens, are under the validator's min_self_delegation.
func isBelowMinSelfDelegation(validator stakingtypes.ValidatorI, shares sdk.Dec) bool {
	return validator.TokensFromShares(shares).TruncateInt().LT(validator.GetMinSelfDelegation())
}

// fastUndelegate undelegates shares from a validator skipping the unbonding period and not creating any unbonding delegations.
func (k Keeper) fastUndelegate(ctx sdk.Context, valAddr sdk.ValAddress, delegator sdk.AccAddress, shares sdk.Dec) (sdkmath.Int, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return sdkmath.Int{}, types.ErrNoDelegatorForAddress
	}

	returnAmount, err := k.stakingKeeper.Unbond(ctx, delegator, valAddr, shares)
	if err != nil {
		return sdkmath.Int{}, err
	}
	returnCoins := sdk.NewCoins(sdk.NewCoin(k.stakingKeeper.BondDenom(ctx), returnAmount))

	// transfer the validator tokens to the not bonded pool
	if validator.IsBonded() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, returnCoins); err != nil {
			panic(err)
		}
	}

	if err := k.bankKeeper.UndelegateCoinsFromModuleToAccount(ctx, stakingtypes.NotBondedPoolName, delegator, returnCoins); err != nil {
		return sdkmath.Int{}, err
	}
	return returnAmount, nil
}

// delegateFromAccount delegates to a validator from an account (vs redelegating from an existing delegation)
func (k Keeper) delegateFromAccount(ctx sdk.Context, valAddr sdk.ValAddress, delegator sdk.AccAddress, amount sdkmath.Int) (sdk.Dec, error) {
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return sdk.Dec{}, types.ErrNoValidatorFound
	}
	// source tokens are from an account, so subtractAccount true and tokenSrc unbonded
	newShares, err := k.stakingKeeper.Delegate(ctx, delegator, amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return sdk.Dec{}, err
	}
	return newShares, nil
}
