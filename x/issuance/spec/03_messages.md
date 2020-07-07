<!--
order: 3
-->

# Messages

The issuer can issue new tokens using a `MsgIssueTokens`

```go
// MsgIssueTokens message type used to issue tokens
type MsgIssueTokens struct {
  Sender   sdk.AccAddress `json:"sender" yaml:"sender"`
  Tokens   sdk.Coin       `json:"tokens" yaml:"tokens"`
  Receiver sdk.AccAddress `json:"receiver" yaml:"receiver"`
}
```

## State Modifications

* New tokens are minted from the issuance module account
* New tokens are transferred from the module account to the receiver

The issuer can redeem (burn) tokens using `MsgRedeemTokens`.

```go
// MsgRedeemTokens message type used to redeem (burn) tokens
type MsgRedeemTokens struct {
  Sender sdk.AccAddress `json:"sender" yaml:"sender"`
  Tokens sdk.Coin       `json:"tokens" yaml:"tokens"`
}
```

## State Modifications

* Tokens are transferred from the owner address to the issuer module account
* Tokens are burned

Addresses can be added to the blocked list using `MsgBlockAddress`

```go
// MsgBlockAddress message type used by the issuer to block an address from holding or transferring tokens
type MsgBlockAddress struct {
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom          string         `json:"denom" yaml:"denom"`
	BlockedAddress sdk.AccAddress `json:"blocked_address" yaml:"blocked_address"`
}
```

## State Modifications

* The address is added to the block list, which prevents the account from holding coins of that denom
* Tokens are sent back to the issuer

The issuer can pause or un-pause the contract using `MsgChangePauseStatus`

```go
// MsgChangePauseStatus message type used by the issuer to issue new tokens
type MsgChangePauseStatus struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Denom  string         `json:"denom" yaml:"denom"`
	Status bool           `json:"status" yaml:"status"`
}
```

## State Modifications

* The `Paused` value of the correspond asset is updated to `Status`.
* Issuance and redemption are paused if `Paused` is false
