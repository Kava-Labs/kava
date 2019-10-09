# State

## Validator Vesting Account type

Validator Vesting Accounts implement the `cosmos-sdk` vesting account spec and extend the `PeriodicVestingAccountType`:

```go
type ValidatorVestingAccount struct {
	*vesting.PeriodicVestingAccount
	ValidatorAddress       sdk.ConsAddress // The validator address which will be used to check if blocks were signed
	ReturnAddress          sdk.AccAddress  `json:"return_address" yaml:"return_address"` // The account where coins will be returned in the event of a failed vesting period
	SigningThreshold       int64           `json:"signing_threshold" yaml:"signing_threshold"` // The percentage of blocks, as an integer between 0 and 100, that must be signed each period for coins to successfully vest.
	MissingSignCount       []int64         `json:"missing_sign_count" yaml:"missing_sign_count"` // An array of two integers which track the number of blocks that were not signed during the current period and the total number of blocks which have passed during the current period, respectively.
	VestingPeriodProgress  [][]int           `json:"vesting_period_progress" yaml:"vesting_period_progress"` //An 2d array with length equal to the number of vesting periods. After each period, the value at the first index of that period is updated with 1 to represent that the period is over. The value at the second index is updated to 0 for unsucessful vesting and 1 for successful vesting.
  DebtAfterFailedVesting sdk.Coins       `json:"debt_after_failed_vesting" yaml:"debt_after_failed_vesting"` // The debt currently owed by the account. Debt accumulates in the event of unsuccessful vesting periods.
}
```

## Stores

There is one `KVStore` in `validator-vesting` which stores
* A mapping from each ValidatorVestingAccount `address` to `[]Byte{0}`
* A mapping from `previous_block_time_prefix` to `time.Time`

The use of `[]Byte{0}` value for each `address` key reflects that this module only accesses the store to get or iterate over keys, and does not require storing an value.
