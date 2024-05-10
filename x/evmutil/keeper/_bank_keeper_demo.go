package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// ConversionMultiplier is the conversion multiplier between akava and ukava
var ScalingFactor = sdkmath.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = ExtendedBankKeeper{}

// ExtendedBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal akava coin on the evm.
// x/evm consumes gas and send coins by minting and burning akava coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the ukava coin and a separate akava balance to manage the
// extra precision needed by the evm.
type ExtendedBankKeeper struct {
	akavaKeeper Keeper
	bk          types.BankKeeper
	ak          types.AccountKeeper
}

// GetBalance returns the total **spendable** balance of akava for a given account by address.
func (k ExtendedBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by ExtendedBankKeeper", EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	integerAmount := spendableCoins.AmountOf(CosmosDenom)
	decimalAmount := k.akavaKeeper.GetBalance(ctx, addr)

	total := integerAmount.Mul(ConversionMultiplier).Add(decimalAmount)

	return sdk.NewCoin(EvmDenom, total)
}

func GetDecimalParts(
	amount sdkmath.Int,
) (sdkmath.Int, sdkmath.Int) {
	integer := amount.Quo(ScalingFactor)
	fractional := amount.Mod(ScalingFactor)

	return integer, fractional
}

func GetDecimalPartDeltas(
	amountBefore sdkmath.Int,
	amountAfter sdkmath.Int,
) (sdkmath.Int, sdkmath.Int) {
	integerBefore, fractionalBefore := GetDecimalParts(amountBefore)
	integerAfter, fractionalAfter := GetDecimalParts(amountAfter)

	// Example: 100.00 -> 99.8
	// integerDelta    = 99 - 100 = -1
	// fractionalDelta = 0.8 - 00 = +0.8
	integerDelta := integerAfter.Sub(integerBefore)
	fractionalDelta := fractionalAfter.Sub(fractionalBefore)

	return integerDelta, fractionalDelta
}

// SendCoins transfers akava coins from a AccAddress to an AccAddress.
func (k ExtendedBankKeeper) SendCoins(
	ctx sdk.Context,
	senderAddr sdk.AccAddress,
	recipientAddr sdk.AccAddress,
	amt sdk.Coins,
) error {
	if err := ValidateCoins(amt); err != nil {
		return err
	}

	// Step 1: Get current sender balance
	senderBal := k.GetBalance(ctx, senderAddr, EvmDenom)

	// Step 2: Check if sender has enough balance
	if senderBal.Amount.LT(amt.AmountOf(EvmDenom)) {
		return errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"insufficient akava balance: %s < %s",
			senderBal, amt,
		)
	}

	// Step 3: Subtract the amount from the sender balance
	newSenderBal := senderBal.Amount.Sub(amt.AmountOf(EvmDenom))

	// Step 4: Deduct the amount from the sender's balance

	// Step 5: Get the recipient balance
	recipientBal := k.GetBalance(ctx, recipientAddr, EvmDenom)

	// Step 6: Add the amount to the recipient balance
	newRecipientBal := recipientBal.Amount.Add(amt.AmountOf(EvmDenom))

}

// SendCoinsFromModuleToAccount transfers akava coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k ExtendedBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	ukava, akava, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if ukava.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneUkavaToAkavaIfNeeded(ctx, senderAddr, akava); err != nil {
		return err
	}

	if err := k.akavaKeeper.SendBalance(ctx, senderAddr, recipientAddr, akava); err != nil {
		return err
	}

	return k.ConvertAkavaToUkava(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers akava coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k ExtendedBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	ukava, akavaNeeded, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if ukava.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneUkavaToAkavaIfNeeded(ctx, senderAddr, akavaNeeded); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.akavaKeeper.SendBalance(ctx, senderAddr, recipientAddr, akavaNeeded); err != nil {
		return err
	}

	return k.ConvertAkavaToUkava(ctx, recipientAddr)
}

// MintCoins mints akava coins by minting the equivalent ukava coins and any remaining akava coins.
// It will panic if the module account does not exist or is unauthorized.
func (k ExtendedBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ukava, akava, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if ukava.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.akavaKeeper.AddBalance(ctx, recipientAddr, akava); err != nil {
		return err
	}

	return k.ConvertAkavaToUkava(ctx, recipientAddr)
}

// BurnCoins burns akava coins by burning the equivalent ukava coins and any remaining akava coins.
// It will panic if the module account does not exist or is unauthorized.
func (k ExtendedBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ukava, akava, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if ukava.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneUkavaToAkavaIfNeeded(ctx, moduleAddr, akava); err != nil {
		return err
	}

	return k.akavaKeeper.RemoveBalance(ctx, moduleAddr, akava)
}

// IsSendEnabledCoins checks the coins provided and returns an ErrSendDisabled
// if any of the coins are not configured for sending. Returns nil if sending is
// enabled for all provided coins.
func (k ExtendedBankKeeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	// IsSendEnabledCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use IsSendEnabledCoins.
	panic("not implemented")
}

// ConvertOneUkavaToAkavaIfNeeded converts 1 ukava to akava for an address if
// its akava balance is smaller than the akavaNeeded amount.
func (k ExtendedBankKeeper) ConvertOneUkavaToAkavaIfNeeded(ctx sdk.Context, addr sdk.AccAddress, akavaNeeded sdkmath.Int) error {
	akavaBal := k.akavaKeeper.GetBalance(ctx, addr)
	if akavaBal.GTE(akavaNeeded) {
		return nil
	}

	ukavaToStore := sdk.NewCoins(sdk.NewCoin(CosmosDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, ukavaToStore); err != nil {
		return err
	}

	// add 1ukava equivalent of akava to addr
	akavaToReceive := ConversionMultiplier
	if err := k.akavaKeeper.AddBalance(ctx, addr, akavaToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertAkavaToUkava converts all available akava to ukava for a given AccAddress.
func (k ExtendedBankKeeper) ConvertAkavaToUkava(ctx sdk.Context, addr sdk.AccAddress) error {
	totalAkava := k.akavaKeeper.GetBalance(ctx, addr)
	ukava, _, err := SplitAkavaCoins(sdk.NewCoins(sdk.NewCoin(EvmDenom, totalAkava)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough akava for a single ukava
	ukavaToReceive := ukava.Amount
	if !ukavaToReceive.IsPositive() {
		return nil
	}

	// remove akava used for converting to ukava
	akavaToBurn := ukavaToReceive.Mul(ConversionMultiplier)
	finalBal := totalAkava.Sub(akavaToBurn)
	if err := k.akavaKeeper.SetBalance(ctx, addr, finalBal); err != nil {
		return err
	}

	fromAddr := k.GetModuleAddress(types.ModuleName)
	if err := k.bk.SendCoins(ctx, fromAddr, addr, sdk.NewCoins(ukava)); err != nil {
		return err
	}

	return nil
}

func (k ExtendedBankKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	addr := k.ak.GetModuleAddress(moduleName)
	if addr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}
	return addr
}

// ValidateEvmCoins validates the coins from evm is valid and is the EvmDenom (akava).
func ValidateCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is akava
	if len(coins) != 1 || coins[0].Denom != EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", EvmDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
