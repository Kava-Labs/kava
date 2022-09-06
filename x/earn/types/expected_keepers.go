package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	SetModuleAccount(sdk.Context, types.ModuleAccountI)
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// DistributionKeeper defines the expected interface needed for community-pool deposits to earn vaults
type DistributionKeeper interface {
	GetFeePool(ctx sdk.Context) (feePool disttypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool disttypes.FeePool)
	GetDistributionAccount(ctx sdk.Context) types.ModuleAccountI
}

// HardKeeper defines the expected interface needed for the hard strategy.
type HardKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error

	GetSyncedDeposit(ctx sdk.Context, depositor sdk.AccAddress) (hardtypes.Deposit, bool)
}

// SavingsKeeper defines the expected interface needed for the savings strategy.
type SavingsKeeper interface {
	Deposit(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error
	Withdraw(ctx sdk.Context, depositor sdk.AccAddress, coins sdk.Coins) error

	GetDeposit(ctx sdk.Context, depositor sdk.AccAddress) (savingstypes.Deposit, bool)
}

// EarnHooks are event hooks called when a user's deposit to a earn vault changes.
type EarnHooks interface {
	AfterVaultDepositCreated(ctx sdk.Context, vaultDenom string, depositor sdk.AccAddress, sharedOwned sdk.Dec)
	BeforeVaultDepositModified(ctx sdk.Context, vaultDenom string, depositor sdk.AccAddress, sharedOwned sdk.Dec)
}

type StakingKeeper interface {
	BondDenom(ctx sdk.Context) (res string)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	ValidateUnbondAmount(
		ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int,
	) (shares sdk.Dec, err error)

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
