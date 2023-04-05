package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

var (
	// ErrInvalidDepositDenom error for invalid deposit denoms
	ErrInvalidDepositDenom = errorsmod.Register(ModuleName, 2, "invalid deposit denom")
	// ErrDepositNotFound error for deposit not found
	ErrDepositNotFound = errorsmod.Register(ModuleName, 3, "deposit not found")
	// ErrInvalidWithdrawAmount error for invalid withdrawal amount
	ErrInvalidWithdrawAmount = errorsmod.Register(ModuleName, 4, "invalid withdrawal amount")
	// ErrInsufficientModAccountBalance error for module account with innsufficient balance
	ErrInsufficientModAccountBalance = errorsmod.Register(ModuleName, 5, "module account has insufficient balance to pay reward")
	// ErrInvalidAccountType error for unsupported accounts
	ErrInvalidAccountType = errorsmod.Register(ModuleName, 6, "receiver account type not supported")
	// ErrAccountNotFound error for accounts that are not found in state
	ErrAccountNotFound = errorsmod.Register(ModuleName, 7, "account not found")
	// ErrInvalidReceiver error for when sending and receiving accounts don't match
	ErrInvalidReceiver = errorsmod.Register(ModuleName, 8, "receiver account must match sender account")
	// ErrMoneyMarketNotFound error for money market param not found
	ErrMoneyMarketNotFound = errorsmod.Register(ModuleName, 9, "no money market found")
	// ErrDepositsNotFound error for no deposits found
	ErrDepositsNotFound = errorsmod.Register(ModuleName, 10, "no deposits found")
	// ErrInsufficientLoanToValue error for when an attempted borrow exceeds maximum loan-to-value
	ErrInsufficientLoanToValue = errorsmod.Register(ModuleName, 11, "not enough collateral supplied by account")
	// ErrMarketNotFound error for when a market for the input denom is not found
	ErrMarketNotFound = errorsmod.Register(ModuleName, 12, "no market found for denom")
	// ErrPriceNotFound error for when a price for the input market is not found
	ErrPriceNotFound = errorsmod.Register(ModuleName, 13, "no price found for market")
	// ErrBorrowExceedsAvailableBalance for when a requested borrow exceeds available module acc balances
	ErrBorrowExceedsAvailableBalance = errorsmod.Register(ModuleName, 14, "exceeds module account balance")
	// ErrBorrowedCoinsNotFound error for when the total amount of borrowed coins cannot be found
	ErrBorrowedCoinsNotFound = errorsmod.Register(ModuleName, 15, "no borrowed coins found")
	// ErrNegativeBorrowedCoins error for when substracting coins from the total borrowed balance results in a negative amount
	ErrNegativeBorrowedCoins = errorsmod.Register(ModuleName, 16, "subtraction results in negative borrow amount")
	// ErrGreaterThanAssetBorrowLimit error for when a proposed borrow would increase borrowed amount over the asset's global borrow limit
	ErrGreaterThanAssetBorrowLimit = errorsmod.Register(ModuleName, 17, "fails global asset borrow limit validation")
	// ErrBorrowEmptyCoins error for when you cannot borrow empty coins
	ErrBorrowEmptyCoins = errorsmod.Register(ModuleName, 18, "cannot borrow zero coins")
	// ErrBorrowNotFound error for when a user's borrow is not found in the store
	ErrBorrowNotFound = errorsmod.Register(ModuleName, 19, "borrow not found")
	// ErrPreviousAccrualTimeNotFound error for no previous accrual time found in store
	ErrPreviousAccrualTimeNotFound = errorsmod.Register(ModuleName, 20, "no previous accrual time found")
	// ErrInsufficientBalanceForRepay error for when requested repay exceeds user's balance
	ErrInsufficientBalanceForRepay = errorsmod.Register(ModuleName, 21, "insufficient balance")
	// ErrBorrowNotLiquidatable error for when a borrow is within valid LTV and cannot be liquidated
	ErrBorrowNotLiquidatable = errorsmod.Register(ModuleName, 22, "borrow not liquidatable")
	// ErrInsufficientCoins error for when there are not enough coins for the operation
	ErrInsufficientCoins = errorsmod.Register(ModuleName, 23, "unrecoverable state - insufficient coins")
	// ErrInsufficientBalanceForBorrow error for when the requested borrow exceeds user's balance
	ErrInsufficientBalanceForBorrow = errorsmod.Register(ModuleName, 24, "insufficient balance")
	// ErrSuppliedCoinsNotFound error for when the total amount of supplied coins cannot be found
	ErrSuppliedCoinsNotFound = errorsmod.Register(ModuleName, 25, "no supplied coins found")
	// ErrNegativeSuppliedCoins error for when substracting coins from the total supplied balance results in a negative amount
	ErrNegativeSuppliedCoins = errorsmod.Register(ModuleName, 26, "subtraction results in negative supplied amount")
	// ErrInvalidWithdrawDenom error for when user attempts to withdraw a non-supplied coin type
	ErrInvalidWithdrawDenom = errorsmod.Register(ModuleName, 27, "no coins of this type deposited")
	// ErrInvalidRepaymentDenom error for when user attempts to repay a non-borrowed coin type
	ErrInvalidRepaymentDenom = errorsmod.Register(ModuleName, 28, "no coins of this type borrowed")
	// ErrInvalidIndexFactorDenom error for when index factor denom cannot be found
	ErrInvalidIndexFactorDenom = errorsmod.Register(ModuleName, 29, "no index factor found for denom")
	// ErrBelowMinimumBorrowValue error for when a proposed borrow position is less than the minimum USD value
	ErrBelowMinimumBorrowValue = errorsmod.Register(ModuleName, 30, "invalid proposed borrow value")
	// ErrExceedsProtocolBorrowableBalance for when a requested borrow exceeds the module account's borrowable balance
	ErrExceedsProtocolBorrowableBalance = errorsmod.Register(ModuleName, 31, "exceeds borrowable module account balance")
	// ErrReservesExceedCash for when the protocol is insolvent because available reserves exceeds available cash
	ErrReservesExceedCash = errorsmod.Register(ModuleName, 32, "insolvency - protocol reserves exceed available cash")
)
