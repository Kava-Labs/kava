package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	earntypes "github.com/kava-labs/kava/x/earn/types"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// ParamSubspace defines the expected Subspace interfacace
type ParamSubspace interface {
	GetParamSet(sdk.Context, paramtypes.ParamSet)
	SetParamSet(sdk.Context, paramtypes.ParamSet)
	WithKeyTable(paramtypes.KeyTable) paramtypes.Subspace
	HasKeyTable() bool
}

// BankKeeper defines the expected interface needed to send coins
type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// StakingKeeper defines the expected staking keeper for module accounts
type StakingKeeper interface {
	GetDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress, maxRetrieve uint16) (delegations []stakingtypes.Delegation)
	GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) (delegations []stakingtypes.Delegation)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	TotalBondedTokens(ctx sdk.Context) sdk.Int
}

// CdpKeeper defines the expected cdp keeper for interacting with cdps
type CdpKeeper interface {
	GetInterestFactor(ctx sdk.Context, collateralType string) (sdk.Dec, bool)
	GetTotalPrincipal(ctx sdk.Context, collateralType string, principalDenom string) (total sdk.Int)
	GetCdpByOwnerAndCollateralType(ctx sdk.Context, owner sdk.AccAddress, collateralType string) (cdptypes.CDP, bool)
	GetCollateral(ctx sdk.Context, collateralType string) (cdptypes.CollateralParam, bool)
}

// HardKeeper defines the expected hard keeper for interacting with Hard protocol
type HardKeeper interface {
	GetDeposit(ctx sdk.Context, depositor sdk.AccAddress) (hardtypes.Deposit, bool)
	GetBorrow(ctx sdk.Context, borrower sdk.AccAddress) (hardtypes.Borrow, bool)

	GetSupplyInterestFactor(ctx sdk.Context, denom string) (sdk.Dec, bool)
	GetBorrowInterestFactor(ctx sdk.Context, denom string) (sdk.Dec, bool)
	GetBorrowedCoins(ctx sdk.Context) (coins sdk.Coins, found bool)
	GetSuppliedCoins(ctx sdk.Context) (coins sdk.Coins, found bool)
}

// SwapKeeper defines the required methods needed by this modules keeper
type SwapKeeper interface {
	GetPoolShares(ctx sdk.Context, poolID string) (shares sdk.Int, found bool)
	GetDepositorSharesAmount(ctx sdk.Context, depositor sdk.AccAddress, poolID string) (shares sdk.Int, found bool)
}

// SavingsKeeper defines the required methods needed by this module's keeper
type SavingsKeeper interface {
	GetDeposit(ctx sdk.Context, depositor sdk.AccAddress) (savingstypes.Deposit, bool)
	GetSavingsModuleAccountBalances(ctx sdk.Context) sdk.Coins
}

// EarnKeeper defines the required methods needed by this modules keeper
type EarnKeeper interface {
	GetVaultTotalShares(ctx sdk.Context, denom string) (shares earntypes.VaultShare, found bool)
	GetVaultAccountShares(ctx sdk.Context, acc sdk.AccAddress) (shares earntypes.VaultShares, found bool)
	IterateVaultRecords(ctx sdk.Context, cb func(record earntypes.VaultRecord) (stop bool))
}

// LiquidKeeper defines the required methods needed by this modules keeper
type LiquidKeeper interface {
	IsDerivativeDenom(ctx sdk.Context, denom string) bool
	GetAllDerivativeDenoms(ctx sdk.Context) (denoms []string)
	GetTotalDerivativeSupply(ctx sdk.Context) sdk.Int
	GetDerivativeSupply(ctx sdk.Context, denom string) sdk.Int
	CollectStakingRewardsByDenom(
		ctx sdk.Context,
		derivativeDenom string,
		destinationModAccount string,
	) (sdk.Coins, error)
}

// AccountKeeper expected interface for the account keeper (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
}

// CDPHooks event hooks for other keepers to run code in response to CDP modifications
type CDPHooks interface {
	AfterCDPCreated(ctx sdk.Context, cdp cdptypes.CDP)
	BeforeCDPModified(ctx sdk.Context, cdp cdptypes.CDP)
}

// HARDHooks event hooks for other keepers to run code in response to HARD modifications
type HARDHooks interface {
	AfterDepositCreated(ctx sdk.Context, deposit hardtypes.Deposit)
	BeforeDepositModified(ctx sdk.Context, deposit hardtypes.Deposit)
	AfterDepositModified(ctx sdk.Context, deposit hardtypes.Deposit)
	AfterBorrowCreated(ctx sdk.Context, borrow hardtypes.Borrow)
	BeforeBorrowModified(ctx sdk.Context, borrow hardtypes.Borrow)
	AfterBorrowModified(ctx sdk.Context, deposit hardtypes.Deposit)
}
