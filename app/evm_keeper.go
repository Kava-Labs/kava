package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// ExpectedEVMBankKeeper the expected interface for the EVM bank keeper wrapper.
type ExpectedEVMBankKeeper interface {
	evmtypes.BankKeeper

	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// Convertion between native gas KAVA (6) to EVM (18).
// 12 decimal difference, so 1_000_000_000_000.
var conversionMultiplier = sdk.NewInt(1_000_000_000_000)

// EVMBankKeeper is a wrapper for bank keeper that converts between EVM (18) and
// native decimals (6).
type EVMBankKeeper struct {
	bankKeeper ExpectedEVMBankKeeper
}

var _ evmtypes.BankKeeper = (*EVMBankKeeper)(nil)

// NewEVMBankKeeper returns a wrapped bank keeper that converts between EVM (18)
// and native decimals (6).
func NewEVMBankKeeper(bk ExpectedEVMBankKeeper) EVMBankKeeper {
	return EVMBankKeeper{
		bankKeeper: bk,
	}
}

// GetBalance returns the 18 decimal **spendable** balance of a specific
// denomination for a given account by address.
func (bk EVMBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	bal := bk.bankKeeper.SpendableCoins(ctx, addr).AmountOf(denom)
	return convertCoinToEvm(sdk.NewCoin(denom, bal))
}

// SendCoinsFromModuleToAccount transfers 18 decimal coins from a ModuleAccount
// to an AccAddress. It will panic if the module account does not exist. An
// error is returned if the recipient address is black-listed or if sending the
// tokens fails.
func (bk EVMBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return bk.bankKeeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, convertCoinsFromEvm(amt))
}

// SendCoinsFromAccountToModule transfers 18 decimal coins from an AccAddress to
// a ModuleAccount. It will panic if the module account does not exist.
func (bk EVMBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return bk.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, convertCoinsFromEvm(amt))
}

// MintCoins creates new 18 decimal coins from thin air and adds it to the
// module account. It will panic if the module account does not exist or is
// unauthorized.
func (bk EVMBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return bk.bankKeeper.MintCoins(ctx, moduleName, convertCoinsFromEvm(amt))
}

// BurnCoins burns 18 decimal coins deletes coins from the balance of the module
// account. It will panic if the module account does not exist or is
// unauthorized.
func (bk EVMBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return bk.bankKeeper.BurnCoins(ctx, moduleName, convertCoinsFromEvm(amt))
}

//____________________________________________________________________________

// convertCoinToEvm converts a sdk.Coin with native decimals to an EVM sdk.Coin
// with 18 decimals.
func convertCoinToEvm(coin sdk.Coin) sdk.Coin {
	// coin.Amount.BigInt() creates a copy of the underlying Int value via Int.Set
	newAmount := sdk.NewIntFromBigInt(coin.Amount.BigInt()).Mul(conversionMultiplier)
	newCoin := sdk.NewCoin(coin.Denom, newAmount)

	return newCoin
}

// convertCoinFromEvm converts an EVM sdk.Coin with 18 decimals back to native 6
// decimal sdk.Coin.
func convertCoinFromEvm(coin sdk.Coin) sdk.Coin {
	// coin.Amount.BigInt() creates a copy of the underlying Int value via Int.Set
	newAmount := sdk.NewIntFromBigInt(coin.Amount.BigInt()).Quo(conversionMultiplier)
	newCoin := sdk.NewCoin(coin.Denom, newAmount)

	if newCoin.IsZero() {
		panic(fmt.Sprintf("EVM coin (%v with 18 decimals) is too small! Conversion from 18 to 6 decimals is zero", coin))
	}

	return newCoin
}

// convertCoinFromEvm converts EVM sdk.Coins with 18 decimals back to native 6
// decimal sdk.Coins.
func convertCoinsFromEvm(coins sdk.Coins) sdk.Coins {
	// Use a new sdk.Coins for deep copy
	var newCoins sdk.Coins

	for _, coin := range coins {
		newCoins = append(newCoins, convertCoinFromEvm(coin))
	}

	return newCoins
}
