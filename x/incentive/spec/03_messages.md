<!--
order: 3
-->

# Messages

Users claim rewards using a `MsgClaimReward`.

```go
// MsgClaimReward message type used to claim rewards
type MsgClaimReward struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom  string         `json:"denom" yaml:"denom"`
}
```

## State Modifications

* Accumulated rewards for active claims are transferred from the `kavadist` module account to the users account as vesting coins
* The corresponding claim object(s) are deleted from the store
