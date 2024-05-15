package types

// IntegerCoinDenom is the denomination for integer coins that are managed by
// x/bank. This is the "true" denomination of the coin, and is also used for
// the reserve to back all fractional coins.
const IntegerCoinDenom = "ukava"

// ExtendedCoinDenom is the denomination for the extended IntegerCoinDenom. This
// not only represents the fractional balance, but the total balance of
// integer + fractional balances.
const ExtendedCoinDenom = "akava"
