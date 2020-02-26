package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// Swap is an interface for handling common actions
type Swap interface {
	GetSwapID() cmn.HexBytes
	GetModuleAccountCoins() sdk.Coins
	Validate() error
}

// AtomicSwap contains the information for an atomic swap
type AtomicSwap struct {
	Swap
	Amount           sdk.Coins      `json:"amount"  yaml:"amount"`
	RandomNumberHash cmn.HexBytes   `json:"random_number_hash"  yaml:"random_number_hash"`
	ExpireHeight     int64          `json:"expire_height"  yaml:"expire_height"`
	Timestamp        int64          `json:"timestamp"  yaml:"timestamp"`
	Sender           sdk.AccAddress `json:"sender"  yaml:"sender"`
	Recipient        sdk.AccAddress `json:"recipient"  yaml:"recipient"`
	SenderOtherChain string         `json:"sender_other_chain"  yaml:"sender_other_chain"`
	ClosedBlock      int64          `json:"closed_block"  yaml:"closed_block"`
	Status           SwapStatus     `json:"status"  yaml:"status"`
}

// GetSwapID calculates the ID of an atomic swap
func (a AtomicSwap) GetSwapID() cmn.HexBytes {
	return CalculateSwapID(a.RandomNumberHash, a.Sender, a.SenderOtherChain)
}

// GetModuleAccountCoins returns the swap's amount as sdk.Coins
func (a AtomicSwap) GetModuleAccountCoins() sdk.Coins {
	return sdk.NewCoins(a.Amount...)
}

// Validate verifies that recipient is not empty
func (a AtomicSwap) Validate() error {
	if len(a.Sender) != AddrByteCount {
		return fmt.Errorf(fmt.Sprintf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Sender)))
	}
	if len(a.Recipient) != AddrByteCount {
		return fmt.Errorf(fmt.Sprintf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Recipient)))
	}
	if len(a.RandomNumberHash) != RandomNumberHashLength {
		return fmt.Errorf(fmt.Sprintf("the length of random number hash should be %d", RandomNumberHashLength))
	}
	if !a.Amount.IsAllPositive() {
		return fmt.Errorf(fmt.Sprintf("the swapped out coin must be positive"))
	}
	return nil
}

// NewAtomicSwap returns a new AtomicSwap
func NewAtomicSwap(amount sdk.Coins, randomNumberHash cmn.HexBytes, expireHeight, timestamp int64, sender,
	recipient sdk.AccAddress, senderOtherChain string, closedBlock int64, status SwapStatus) AtomicSwap {
	return AtomicSwap{
		Amount:           amount,
		RandomNumberHash: randomNumberHash,
		ExpireHeight:     expireHeight,
		Timestamp:        timestamp,
		Sender:           sender,
		Recipient:        recipient,
		SenderOtherChain: senderOtherChain,
		ClosedBlock:      closedBlock,
		Status:           status,
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
