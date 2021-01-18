<!--
order: 3
-->

# Messages

There are three messages in the hard module. Deposit allows users to deposit assets to the hard module. In version 2, depositors will be able to use their deposits as collateral to borrow from hard. Withdraw removes assets from the hard module, returning them to the user. Claim allows users to claim earned HARD tokens.

```go
// MsgDeposit deposit asset to the hard module.
type MsgDeposit struct {
  Depositor   sdk.AccAddress `json:"depositor" yaml:"depositor"`
  Amount      sdk.Coin       `json:"amount" yaml:"amount"`
}

// MsgWithdraw withdraw from the hard module.
type MsgWithdraw struct {
  Depositor   sdk.AccAddress `json:"depositor" yaml:"depositor"`
  Amount      sdk.Coin       `json:"amount" yaml:"amount"`
}

// MsgClaimReward message type used to claim HARD tokens
type MsgClaimReward struct {
  Sender           sdk.AccAddress `json:"sender" yaml:"sender"`
  Receiver         sdk.AccAddress `json:"receiver" yaml:"receiver"`
  DepositDenom     string         `json:"deposit_denom" yaml:"deposit_denom"`
  RewardMultiplier string         `json:"reward_multiplier" yaml:"reward_multiplier"`
  DepositType      string         `json:"deposit_type" yaml:"deposit_type"`
}
```
