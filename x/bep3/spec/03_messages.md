<!--
order: 3
-->

# Messages

## Create swap

Swaps are created using the `MsgCreateAtomicSwap` message type.

```go
// MsgCreateAtomicSwap contains an AtomicSwap struct
type MsgCreateAtomicSwap struct {
	From                sdk.AccAddress   `json:"from"  yaml:"from"`
	To                  sdk.AccAddress   `json:"to"  yaml:"to"`
	RecipientOtherChain string           `json:"recipient_other_chain"  yaml:"recipient_other_chain"`
	SenderOtherChain    string           `json:"sender_other_chain"  yaml:"sender_other_chain"`
	RandomNumberHash    tmbytes.HexBytes `json:"random_number_hash"  yaml:"random_number_hash"`
	Timestamp           int64            `json:"timestamp"  yaml:"timestamp"`
	Amount              sdk.Coins        `json:"amount"  yaml:"amount"`
	HeightSpan          int64            `json:"height_span"  yaml:"height_span"`
}
```

## Claim swap

Active swaps are claimed using the `MsgClaimAtomicSwap` message type.

```go
// MsgClaimAtomicSwap defines a AtomicSwap claim
type MsgClaimAtomicSwap struct {
	From         sdk.AccAddress   `json:"from"  yaml:"from"`
	SwapID       tmbytes.HexBytes `json:"swap_id"  yaml:"swap_id"`
	RandomNumber tmbytes.HexBytes `json:"random_number"  yaml:"random_number"`
}
```

## Refund swap

Expired swaps are refunded using the `MsgRefundAtomicSwap` message type.

```go
// MsgRefundAtomicSwap defines a refund msg
type MsgRefundAtomicSwap struct {
	From   sdk.AccAddress   `json:"from" yaml:"from"`
	SwapID tmbytes.HexBytes `json:"swap_id" yaml:"swap_id"`
}
```