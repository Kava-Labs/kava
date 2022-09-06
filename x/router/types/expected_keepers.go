package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
)

type StakingKeeper interface {
	BondDenom(ctx sdk.Context) (res string)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)

	Delegate(
		ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc stakingtypes.BondStatus,
		validator stakingtypes.Validator, subtractAccount bool,
	) (newShares sdk.Dec, err error)
	Undelegate(
		ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec,
	) (time.Time, error)
}

type LiquidKeeper interface {
	TokenToDerivative(ctx sdk.Context, valAddr sdk.ValAddress, amount sdk.Int) (sdk.Coin, error)
	MintDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Coin, error)
	BurnDerivative(ctx sdk.Context, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) (sdk.Dec, error)
}

type EarnKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, amount sdk.Coin, depositStrategy earntypes.StrategyType) error
	Withdraw(ctx sdk.Context, from sdk.AccAddress, wantAmount sdk.Coin, withdrawStrategy earntypes.StrategyType) error
}
