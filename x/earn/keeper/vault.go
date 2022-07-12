package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Vault interface {
	// The denom of the coin supported by this vault. Only this coin can be
	// deposited and withdrawn.
	GetDenom() string

	// GetTotalSupplied returns the total balance supplied to this vault. This
	// is not the current value of the vault, but the count of GetDenom() coins
	// that were originally supplied.
	GetTotalSupplied() (sdk.Coin, error)

	// GetTotalValue returns the total **value** of all coins in this vault,
	// i.e. the realizable total value denominated by GetDenom() if the vault
	// were to liquidate its entire strategies.
	GetTotalValue() (sdk.Coin, error)

	// GetAccountSupplied returns the supplied amount for a single address within this vault.
	GetAccountSupplied(acc sdk.AccAddress) (sdk.Coin, error)

	// GetAccountValue returns the value of a single address within this vault
	// if the account were to withdraw their entire balance.
	GetAccountValue(acc sdk.AccAddress) (sdk.Coin, error)

	// GetStrategy returns the strategy this vault is associated with.
	GetStrategy() Strategy

	// Deposit a coin into this vault. The coin denom must match GetDenom().
	Deposit(acc sdk.AccAddress, amount sdk.Coin) error

	// Withdraw a coin from this vault. The coin denom must match GetDenom().
	Withdraw(acc sdk.AccAddress, amount sdk.Coin) error
}

// TODO: Convert to proto message, to be marshalled and stored in state
type BaseVault struct {
	denom       string
	strategy    Strategy
	totalSupply sdk.Int
}

var _ Vault = (*BaseVault)(nil)

func NewBaseVault(denom string, strategy Strategy) *BaseVault {
	return &BaseVault{
		denom:    denom,
		strategy: strategy,
	}
}

func (v *BaseVault) GetDenom() string {
	return v.denom
}

func (v *BaseVault) GetTotalSupplied() (sdk.Coin, error) {
	return sdk.Coin{}, nil
}

func (v *BaseVault) GetTotalValue() (sdk.Coin, error) {
	return v.GetStrategy().GetEstimatedTotalAssets(v.GetDenom())
}

func (v *BaseVault) GetAccountSupplied(acc sdk.AccAddress) (sdk.Coin, error) {

	return sdk.Coin{}, nil
}

func (v *BaseVault) GetAccountValue(acc sdk.AccAddress) (sdk.Coin, error) {
	return sdk.Coin{}, nil
}

func (v *BaseVault) GetStrategy() Strategy {
	return v.strategy
}

func (v *BaseVault) Deposit(acc sdk.AccAddress, amount sdk.Coin) error {

	return nil
}

func (v *BaseVault) Withdraw(acc sdk.AccAddress, amount sdk.Coin) error {

	return nil
}
