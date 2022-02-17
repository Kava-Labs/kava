package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/evmutils/types"
)

const (
	// EvmDenom is the gas denom used by evm
	EvmDenom = "akava"

	// CosmosDenom is the gas denom used by the kava app
	CosmosDenom = "ukava"
)

// ConversionMultiplier is the conversion multiplier between akava and ukava
var ConversionMultiplier = sdk.NewInt(1_000_000_000_000)

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal akava coin on the evm.
// This keeper uses both the ukava coin and a separate akava coin to
// manage the extra percision needed by the evm.
type EvmBankKeeper struct {
	bk types.BankKeeper
}

func NewEvmBankKeeper(bk types.BankKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		bk: bk,
	}
}

// GetBalance returns the total **spendable** balance of akava for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	ukava := spendableCoins.AmountOf(CosmosDenom)
	akava := spendableCoins.AmountOf(EvmDenom)
	total := ukava.Mul(ConversionMultiplier).Add(akava)
	return sdk.NewCoin(EvmDenom, total)
}

// SendCoinsFromModuleToAccount transfers 18 decimal coins from a ModuleAccount
// to an AccAddress. It will panic if the module account does not exist. An
// error is returned if the recipient address is black-listed or if sending the
// tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	coins, err := splitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, coins); err != nil {
		return err
	}

	return k.convertAkavaToUkava(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers akava coins from an AccAddress to
// a ModuleAccount. It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	coins, err := splitAkavaCoins(amt)
	if err != nil {
		return err
	}

	// convert 1 ukava to akava if sender account does not have enough akava to cover
	akavaNeeded := coins.AmountOf(EvmDenom)
	akavaBal := k.bk.GetBalance(ctx, senderAddr, EvmDenom).Amount
	if akavaBal.LT(akavaNeeded) {
		if err := k.convertOneUkavaToAkava(ctx, senderAddr); err != nil {
			return err
		}
	}

	return k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, coins)
}

// MintCoins mints akava coins by minting the equivalent ukava coins and any remaining akava coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	coins, err := splitAkavaCoins(amt)
	if err != nil {
		return err
	}

	return k.bk.MintCoins(ctx, moduleName, coins)
}

// BurnCoins burns akava coins by burning the equivalent ukava coins and any remaining akava coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	coins, err := splitAkavaCoins(amt)
	if err != nil {
		return err
	}
	return k.bk.BurnCoins(ctx, moduleName, coins)
}

// validateEvmCoins validates that the coins coming from evm is valid and the required denom (akava)
func validateEvmCoins(coins sdk.Coins) error {
	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that we don't have non akava denoms
	if len(coins) > 1 || coins[0].Denom != EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", EvmDenom)
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}

// convertOneUkavaToAkava converts 1 ukava to akava for a given account by address.
func (k EvmBankKeeper) convertOneUkavaToAkava(ctx sdk.Context, addr sdk.AccAddress) error {
	coinsToBurn := sdk.NewCoins(sdk.NewCoin(CosmosDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, coinsToBurn); err != nil {
		return err
	}
	if err := k.bk.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
		return err
	}

	coinsToMint := sdk.NewCoins(sdk.NewCoin(EvmDenom, ConversionMultiplier))
	if err := k.bk.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return err
	}

	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coinsToMint); err != nil {
		return err
	}

	return nil
}

// convertAkavaToUkava converts all available akava to ukava for a given account by address.
func (k EvmBankKeeper) convertAkavaToUkava(ctx sdk.Context, addr sdk.AccAddress) error {
	akava := k.bk.GetBalance(ctx, addr, EvmDenom)
	coins, err := splitAkavaCoins(sdk.NewCoins(akava))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough akava for a single ukava
	ukavaToReceive := coins.AmountOf(CosmosDenom)
	if !ukavaToReceive.IsPositive() {
		return nil
	}

	coinsToBurn := sdk.NewCoins(sdk.NewCoin(EvmDenom, ukavaToReceive.Mul(ConversionMultiplier)))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, coinsToBurn); err != nil {
		return err
	}

	if err := k.bk.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
		return err
	}

	coinsToMint := sdk.NewCoins(sdk.NewCoin(CosmosDenom, ukavaToReceive))
	if err := k.bk.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return err
	}

	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coinsToMint); err != nil {
		return err
	}

	return nil
}

// splitAkavaCoins splits akava coins to the equivalent ukava coins and any remaining akava coins.
// An error will be returned if the coins are not valid or if the coins are not the akava denom.
func splitAkavaCoins(coins sdk.Coins) (sdk.Coins, error) {
	splitCoins := sdk.Coins{}

	if len(coins) == 0 {
		return coins, nil
	}

	if err := validateEvmCoins(coins); err != nil {
		return splitCoins, err
	}

	// split the akava coins into the  ukava and the remaining akava
	// note: we should always have 1 coin here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		splitCoins = append(splitCoins, sdk.NewCoin(EvmDenom, remainingBalance))
	}
	ukavaAmount := coin.Amount.Quo(ConversionMultiplier)
	if ukavaAmount.IsPositive() {
		splitCoins = append(splitCoins, sdk.NewCoin(CosmosDenom, ukavaAmount))
	}

	splitCoins.Sort()

	return splitCoins, nil
}
