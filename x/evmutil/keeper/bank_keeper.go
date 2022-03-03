package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

const (
	// EvmDenom is the gas denom used by the evm
	EvmDenom = "akava"

	// CosmosDenom is the gas denom used by the kava app
	CosmosDenom = "ukava"
)

// ConversionMultiplier is the conversion multiplier between akava and ukava
var ConversionMultiplier = sdk.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal akava coin on the evm.
// x/evm consumes gas and send coins by minting and burning akava coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the ukava coin and a separate akava balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	akavaKeeper Keeper
	bk          types.BankKeeper
	ak          types.AccountKeeper
}

func NewEvmBankKeeper(akavaKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		akavaKeeper: akavaKeeper,
		bk:          bk,
		ak:          ak,
	}
}

// GetBalance returns the total **spendable** balance of akava for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	ukava := spendableCoins.AmountOf(CosmosDenom)
	akava := k.akavaKeeper.GetBalance(ctx, addr)
	total := ukava.Mul(ConversionMultiplier).Add(akava)
	return sdk.NewCoin(EvmDenom, total)
}

// SendCoinsFromModuleToAccount transfers akava coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	ukava, akava, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if ukava.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	if err := k.akavaKeeper.SendBalance(ctx, k.GetModuleAddress(senderModule), recipientAddr, akava); err != nil {
		return err
	}

	return k.ConvertAkavaToUkava(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers akava coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	ukava, akavaNeeded, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	// convert 1 ukava to akava if sender account does not have enough akava to cover
	akavaBal := k.akavaKeeper.GetBalance(ctx, senderAddr)
	if akavaBal.LT(akavaNeeded) {
		if err := k.ConvertOneUkavaToAkava(ctx, senderAddr); err != nil {
			return err
		}
	}

	if ukava.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	return k.akavaKeeper.SendBalance(ctx, senderAddr, k.GetModuleAddress(recipientModule), akavaNeeded)
}

// MintCoins mints akava coins by minting the equivalent ukava coins and any remaining akava coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ukava, akava, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if ukava.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	return k.akavaKeeper.AddBalance(ctx, k.GetModuleAddress(moduleName), akava)
}

// BurnCoins burns akava coins by burning the equivalent ukava coins and any remaining akava coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ukava, akava, err := SplitAkavaCoins(amt)
	if err != nil {
		return err
	}

	if ukava.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(ukava)); err != nil {
			return err
		}
	}

	return k.akavaKeeper.RemoveBalance(ctx, k.GetModuleAddress(moduleName), akava)
}

// ConvertOneUkavaToAkava converts 1 ukava to akava for a given AccAddress.
func (k EvmBankKeeper) ConvertOneUkavaToAkava(ctx sdk.Context, addr sdk.AccAddress) error {
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
func (k EvmBankKeeper) ConvertAkavaToUkava(ctx sdk.Context, addr sdk.AccAddress) error {
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

	if err := k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(ukava)); err != nil {
		return err
	}

	return nil
}

func (k EvmBankKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	addr := k.ak.GetModuleAddress(moduleName)
	if addr == nil {
		panic(sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}
	return addr
}

// SplitAkavaCoins splits akava coins to the equivalent ukava coins and any remaining akava balance.
// An error will be returned if the coins are not valid or if the coins are not the akava denom.
func SplitAkavaCoins(coins sdk.Coins) (sdk.Coin, sdk.Int, error) {
	akava := sdk.ZeroInt()
	ukava := sdk.NewCoin(CosmosDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return ukava, akava, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return ukava, akava, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		akava = remainingBalance
	}
	ukavaAmount := coin.Amount.Quo(ConversionMultiplier)
	if ukavaAmount.IsPositive() {
		ukava = sdk.NewCoin(CosmosDenom, ukavaAmount)
	}

	return ukava, akava, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the EvmDenom (akava).
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is akava
	if len(coins) != 1 || coins[0].Denom != EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", EvmDenom)
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
