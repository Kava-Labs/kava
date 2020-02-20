package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmm "github.com/tendermint/tendermint/libs/common"
)

// HTLT is an Hash-Time Locked Transaction on Kava
// TODO: model after AtomicSwap?
type HTLT struct {
	SwapID              cmm.HexBytes   `json:"swap_id"`
	From                sdk.AccAddress `json:"from"`
	To                  sdk.AccAddress `json:"to"`
	RecipientOtherChain string         `json:"recipient_other_chain"`
	SenderOtherChain    string         `json:"sender_other_chain"`
	RandomNumberHash    cmm.HexBytes   `json:"random_number_hash"`
	Timestamp           int64          `json:"timestamp"`
	Amount              sdk.Coins      `json:"amount"`
	ExpectedIncome      string         `json:"expected_income"`
	HeightSpan          int64          `json:"height_span"`
	CrossChain          bool           `json:"cross_chain"`
	ExpirationBlock     uint64         `json:"expiration_block"`
}

// NewHTLT returns a new HTLT
func NewHTLT(swapID cmm.HexBytes, from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash cmm.HexBytes, timestamp int64, amount sdk.Coins,
	expectedIncome string, heightSpan int64, crossChain bool, expirationBlock uint64) HTLT {
	return HTLT{
		SwapID:              swapID,
		From:                from,
		To:                  to,
		RecipientOtherChain: recipientOtherChain,
		SenderOtherChain:    senderOtherChain,
		RandomNumberHash:    randomNumberHash,
		Timestamp:           timestamp,
		Amount:              amount,
		ExpectedIncome:      expectedIncome,
		HeightSpan:          heightSpan,
		CrossChain:          crossChain,
		ExpirationBlock:     expirationBlock,
	}
}

// HTLTs is a slice of HTLT
type HTLTs []HTLT
