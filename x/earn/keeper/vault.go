package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/earn/types"
)

// GetVaultTotalSupplied returns the total balance supplied to the vault. This
// may not necessarily be the current value of the vault, as it is the sum
// of the supplied denom and the value may be higher due to accumulated APYs.
func (k *Keeper) GetVaultTotalShares(
	ctx sdk.Context,
	denom string,
) (types.VaultShare, bool) {
	vault, found := k.GetVaultRecord(ctx, denom)
	if !found {
		return types.VaultShare{}, false
	}

	return vault.TotalShares, true
}

// GetTotalValue returns the total **value** of all coins in this vault,
// i.e. the realizable total value denominated by GetDenom() if the vault
// were to liquidate its entire strategies.
//
// **Note:** This does not include the tokens held in bank by the module
// account. If it were to be included, also note that the module account is
// unblocked and can receive funds from bank sends.
func (k *Keeper) GetVaultTotalValue(
	ctx sdk.Context,
	denom string,
) (sdk.Coin, error) {
	enabledVault, found := k.GetAllowedVault(ctx, denom)
	if !found {
		return sdk.Coin{}, types.ErrVaultRecordNotFound
	}

	strategy, err := k.GetStrategy(enabledVault.Strategies[0])
	if err != nil {
		return sdk.Coin{}, types.ErrInvalidVaultStrategy
	}

	return strategy.GetEstimatedTotalAssets(ctx, enabledVault.Denom)
}

// GetVaultAccountSupplied returns the supplied amount for a single address
// within a vault.
func (k *Keeper) GetVaultAccountShares(
	ctx sdk.Context,
	acc sdk.AccAddress,
) (types.VaultShares, bool) {
	vaultShareRecord, found := k.GetVaultShareRecord(ctx, acc)
	if !found {
		return nil, false
	}

	return vaultShareRecord.Shares, true
}

// GetVaultAccountValue returns the value of a single address within a vault
// if the account were to withdraw their entire balance.
func (k *Keeper) GetVaultAccountValue(
	ctx sdk.Context,
	denom string,
	acc sdk.AccAddress,
) (sdk.Coin, error) {
	accShares, found := k.GetVaultAccountShares(ctx, acc)
	if !found {
		return sdk.Coin{}, fmt.Errorf("account vault share record for %s not found", denom)
	}

	return k.ConvertToAssets(ctx, accShares.GetShare(denom))
}
