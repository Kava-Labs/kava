<!--
order: 3
-->

# Messages


MsgDeposit adds liquidity to a pool:

```go
// MsgDeposit deposits liquidity into a pool
type MsgDeposit struct {
	Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
	TokenA    sdk.Coin       `json:"token_a" yaml:"token_a"`
	TokenB    sdk.Coin       `json:"token_b" yaml:"token_b"`
	Slippage  sdk.Dec        `json:"slippage" yaml:"slippage"`
	Deadline  int64          `json:"deadline" yaml:"deadline"`
}
```

The first deposit to a pool results in a `PoolRecord` being created. For each deposit, a `ShareRecord` is created or updated, depending on if the depositor has an existing deposit. The deposited tokens are converted to shares. For the first deposit to a pool, shares are equal to the geometric mean of the deposited amount. For example, depositing 200 TokenA and 100 TokenB will create `sqrt(100 * 200) = 141` shares. For subsequent deposits, shares are issued equal to the current conversion between tokens and shares in that pool.

MsgWithdraw removes liquidity from a pool:

```go
// MsgWithdraw deposits liquidity into a pool
type MsgWithdraw struct {
	From      sdk.AccAddress `json:"from" yaml:"from"`
	Shares    sdkmath.Int        `json:"shares" yaml:"shares"`
	MinTokenA sdk.Coin       `json:"min_token_a" yaml:"min_token_a"`
	MinTokenB sdk.Coin       `json:"min_token_b" yaml:"min_token_b"`
	Deadline  int64          `json:"deadline" yaml:"deadline"`
}
```
When withdrawing from a pool, the user specifies the amount of shares they want to withdraw, as well as the minimum amount of tokenA and tokenB that they must receive for the transaction to succeed. When withdrawing, the `ShareRecord` of the user will be decremented by the corresponding amount of shares, or deleted in the case that all liquidity has been withdrawn. If all shares of a pool have been withdrawn from a pool, the `PoolRecord` will be deleted.

MsgSwapExactForTokens trades an exact amount of input tokens for a variable amount of output tokens, with a specified maximum slippage tolerance.

```go
// MsgSwapExactForTokens trades an exact coinA for coinB
type MsgSwapExactForTokens struct {
	Requester   sdk.AccAddress `json:"requester" yaml:"requester"`
	ExactTokenA sdk.Coin       `json:"exact_token_a" yaml:"exact_token_a"`
	TokenB      sdk.Coin       `json:"token_b" yaml:"token_b"`
	Slippage    sdk.Dec        `json:"slippage" yaml:"slippage"`
	Deadline    int64          `json:"deadline" yaml:"deadline"`
}
```

When trading exact inputs for variable outputs, the swap fee is removed from TokenA and added to the pool, then slippage is calculated based on the actual amount of TokenB received compared to the desired amount of TokenB. If the realized slippage of the trade is greater than the specified slippage tolerance, the transaction fails.

MsgSwapForExactTokens trades a variable amount of input tokens for an exact amount of output tokens, with a specified maximum slippage tolerance.

```go
// MsgSwapForExactTokens trades coinA for an exact coinB
type MsgSwapForExactTokens struct {
	Requester   sdk.AccAddress `json:"requester" yaml:"requester"`
	TokenA      sdk.Coin       `json:"token_a" yaml:"token_a"`
	ExactTokenB sdk.Coin       `json:"exact_token_b" yaml:"exact_token_b"`
	Slippage    sdk.Dec        `json:"slippage" yaml:"slippage"`
	Deadline    int64          `json:"deadline" yaml:"deadline"`
}
```

When trading variable inputs for exact outputs, the fee swap fee is removed from TokenA and added to the pool, then slippage is calculated based on the actual amount of TokenA required to acquire the exact TokenB amount versus the desired TokenA required. If the realized slippage of the trade is greater than the specified slippage tolerance, the transaction fails.
