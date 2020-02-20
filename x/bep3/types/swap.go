package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmm "github.com/tendermint/tendermint/libs/common"
)

// AtomicSwap is an Hash-Time Locked Transaction (HTLT) on Kava
type AtomicSwap struct {
	SwapID              cmm.HexBytes   `json:"swap_id"`
	From                sdk.AccAddress `json:"from"`
	To                  sdk.AccAddress `json:"to"`
	RecipientOtherChain string         `json:"recipient_other_chain"`
	SenderOtherChain    string         `json:"sender_other_chain"`
	RandomNumberHash    cmm.HexBytes   `json:"random_number_hash"`
	Timestamp           int64          `json:"timestamp"`
	Amount              sdk.Coins      `json:"amount"`
	ExpectedIncome      string         `json:"expected_income"`
	CrossChain          bool           `json:"cross_chain"`
	ExpirationBlock     uint64         `json:"expiration_block"`
}

// NewAtomicSwap returns a new AtomicSwap
func NewAtomicSwap(swapID cmm.HexBytes, from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash cmm.HexBytes, timestamp int64, amount sdk.Coins,
	expectedIncome string, crossChain bool, expirationBlock uint64) AtomicSwap {
	return AtomicSwap{
		SwapID:              swapID,
		From:                from,
		To:                  to,
		RecipientOtherChain: recipientOtherChain,
		SenderOtherChain:    senderOtherChain,
		RandomNumberHash:    randomNumberHash,
		Timestamp:           timestamp,
		Amount:              amount,
		ExpectedIncome:      expectedIncome,
		CrossChain:          crossChain,
		ExpirationBlock:     expirationBlock,
	}
}

// AtomicSwaps is a slice of AtomicSwap
type AtomicSwaps []AtomicSwap

// SwapStatus is the status of an AtomicSwap
type SwapStatus byte

const (
	NULL      SwapStatus = 0x00
	Open      SwapStatus = 0x01
	Completed SwapStatus = 0x02
	Expired   SwapStatus = 0x03
)

// NewSwapStatusFromString converts string to SwapStatus type
func NewSwapStatusFromString(str string) SwapStatus {
	switch str {
	case "Open", "open":
		return Open
	case "Completed", "completed":
		return Completed
	case "Expired", "expired":
		return Expired
	default:
		return NULL
	}
}

// String returns the string representation of a SwapStatus
func (status SwapStatus) String() string {
	switch status {
	case Open:
		return "Open"
	case Completed:
		return "Completed"
	case Expired:
		return "Expired"
	default:
		return "NULL"
	}
}

// MarshalJSON marshals the SwapStatus
func (status SwapStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// UnmarshalJSON unmarshals the SwapStatus
func (status *SwapStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*status = NewSwapStatusFromString(s)
	return nil
}
