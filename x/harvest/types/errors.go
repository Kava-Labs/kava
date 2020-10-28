package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DONTCOVER

var (
	// ErrInvalidDepositDenom error for invalid deposit denoms
	ErrInvalidDepositDenom = sdkerrors.Register(ModuleName, 2, "invalid deposit denom")
	// ErrDepositNotFound error for deposit not found
	ErrDepositNotFound = sdkerrors.Register(ModuleName, 3, "deposit not found")
	// ErrInvaliWithdrawAmount error for invalid withdrawal amount
	ErrInvaliWithdrawAmount = sdkerrors.Register(ModuleName, 4, "withdrawal amount exceeds deposit amount")
	// ErrInvalidDepositType error for invalid deposit type
	ErrInvalidDepositType = sdkerrors.Register(ModuleName, 5, "invalid deposit type")
	// ErrClaimNotFound error for claim not found
	ErrClaimNotFound = sdkerrors.Register(ModuleName, 6, "claim not found")
	// ErrZeroClaim error for claim amount rounded to zero
	ErrZeroClaim = sdkerrors.Register(ModuleName, 7, "cannot claim - claim amount rounds to zero")
	// ErrLPScheduleNotFound error for liquidity provider rewards schedule not found
	ErrLPScheduleNotFound = sdkerrors.Register(ModuleName, 8, "no liquidity provider rewards schedule found")
	// ErrGovScheduleNotFound error for governance distribution rewards schedule not found
	ErrGovScheduleNotFound = sdkerrors.Register(ModuleName, 9, "no governance rewards schedule found")
	// ErrInvalidMultiplier error for multiplier not found
	ErrInvalidMultiplier = sdkerrors.Register(ModuleName, 10, "invalid rewards multiplier")
	// ErrInsufficientModAccountBalance error for module account with innsufficient balance
	ErrInsufficientModAccountBalance = sdkerrors.Register(ModuleName, 11, "module account has insufficient balance to pay reward")
	// ErrInvalidAccountType error for unsupported accounts
	ErrInvalidAccountType = sdkerrors.Register(ModuleName, 12, "receiver account type not supported")
	// ErrAccountNotFound error for accounts that are not found in state
	ErrAccountNotFound = sdkerrors.Register(ModuleName, 13, "account not found")
	// ErrClaimExpired error for expired claims
	ErrClaimExpired = sdkerrors.Register(ModuleName, 14, "claim period expired")
	// ErrInvalidReceiver error for when sending and receiving accounts don't match
	ErrInvalidReceiver = sdkerrors.Register(ModuleName, 15, "receiver account must match sender account")
	// ErrMoneyMarketNotFound error for money market param not found
	ErrMoneyMarketNotFound = sdkerrors.Register(ModuleName, 16, "no money market found")
	// ErrDepositsNotFound error for no deposits found
	ErrDepositsNotFound = sdkerrors.Register(ModuleName, 17, "no deposits found")
	// ErrInsufficientLoanToValue error for when an attempted borrow exceeds maximum loan-to-value
	ErrInsufficientLoanToValue = sdkerrors.Register(ModuleName, 18, "total deposited value is insufficient for borrow request")
	// ErrMarketNotFound error for when a market for the input denom is not found
	ErrMarketNotFound = sdkerrors.Register(ModuleName, 19, "no market found for denom")
)
