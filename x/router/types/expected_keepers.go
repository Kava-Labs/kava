package types

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
)

type StakingKeeper interface {
	BondDenom(ctx context.Context) (res string, err error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)

	Delegate(
		ctx context.Context, delAddr sdk.AccAddress, bondAmt sdkmath.Int, tokenSrc stakingtypes.BondStatus,
		validator stakingtypes.Validator, subtractAccount bool,
	) (newShares sdkmath.LegacyDec, err error)
	Undelegate(
		ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdkmath.LegacyDec,
	) (time.Time, sdkmath.Int, error)
}

type LiquidKeeper interface {
	DerivativeFromTokens(ctx sdk.Context, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Coin, error)
	MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Coin, error)
	BurnDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdkmath.LegacyDec, error)
}

type EarnKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coin, depositStrategy earntypes.StrategyType) error
	Withdraw(ctx sdk.Context, from sdk.AccAddress, wantAmount sdk.Coin, withdrawStrategy earntypes.StrategyType) (sdk.Coin, error)
}
