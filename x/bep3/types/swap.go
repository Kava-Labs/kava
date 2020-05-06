package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AtomicSwap contains the information for an atomic swap
type AtomicSwap struct {
	Amount              sdk.Coins        `json:"amount"  yaml:"amount"`
	RandomNumberHash    tmbytes.HexBytes `json:"random_number_hash"  yaml:"random_number_hash"`
	ExpireHeight        int64            `json:"expire_height"  yaml:"expire_height"`
	Timestamp           int64            `json:"timestamp"  yaml:"timestamp"`
	Sender              sdk.AccAddress   `json:"sender"  yaml:"sender"`
	Recipient           sdk.AccAddress   `json:"recipient"  yaml:"recipient"`
	SenderOtherChain    string           `json:"sender_other_chain"  yaml:"sender_other_chain"`
	RecipientOtherChain string           `json:"recipient_other_chain"  yaml:"recipient_other_chain"`
	ClosedBlock         int64            `json:"closed_block"  yaml:"closed_block"`
	Status              SwapStatus       `json:"status"  yaml:"status"`
	CrossChain          bool             `json:"cross_chain"  yaml:"cross_chain"`
	Direction           SwapDirection    `json:"direction"  yaml:"direction"`
}

// NewAtomicSwap returns a new AtomicSwap
func NewAtomicSwap(amount sdk.Coins, randomNumberHash tmbytes.HexBytes, expireHeight, timestamp int64, sender,
	recipient sdk.AccAddress, senderOtherChain string, recipientOtherChain string, closedBlock int64,
	status SwapStatus, crossChain bool, direction SwapDirection) AtomicSwap {
	return AtomicSwap{
		Amount:              amount,
		RandomNumberHash:    randomNumberHash,
		ExpireHeight:        expireHeight,
		Timestamp:           timestamp,
		Sender:              sender,
		Recipient:           recipient,
		SenderOtherChain:    senderOtherChain,
		RecipientOtherChain: recipientOtherChain,
		ClosedBlock:         closedBlock,
		Status:              status,
		CrossChain:          crossChain,
		Direction:           direction,
	}
}

// GetSwapID calculates the ID of an atomic swap
func (a AtomicSwap) GetSwapID() tmbytes.HexBytes {
	return CalculateSwapID(a.RandomNumberHash, a.Sender, a.SenderOtherChain)
}

// GetCoins returns the swap's amount as sdk.Coins
func (a AtomicSwap) GetCoins() sdk.Coins {
	return sdk.NewCoins(a.Amount...)
}

// Validate verifies that recipient is not empty
func (a AtomicSwap) Validate() error {
	if !a.Amount.IsValid() {
		return fmt.Errorf("invalid amount: %s", a.Amount)
	}
	if !a.Amount.IsAllPositive() {
		return fmt.Errorf("the swapped out coin must be positive: %s", a.Amount)
	}
	if len(a.RandomNumberHash) != RandomNumberHashLength {
		return fmt.Errorf("the length of random number hash should be %d", RandomNumberHashLength)
	}
	if a.ExpireHeight == 0 {
		return errors.New("expire height cannot be 0")
	}
	if a.Timestamp == 0 {
		return errors.New("timestamp cannot be 0")
	}
	if a.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender cannot be empty")
	}
	if a.Recipient.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient cannot be empty")
	}
	if len(a.Sender) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Sender))
	}
	if len(a.Recipient) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Recipient))
	}
	// NOTE: we don't validate from bech32 because we don't know the prefix
	if strings.TrimSpace(a.SenderOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender other chain cannot be blank")
	}
	if strings.TrimSpace(a.RecipientOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient other chain cannot be blank")
	}
	if a.ClosedBlock == 0 {
		return errors.New("closed block cannot be 0")
	}
	if a.Status == NULL {
		return errors.New("swap status cannot be nil")
	}
	if a.Direction == INVALID {
		return errors.New("invalid swap direction")
	}
	return nil
}

// String implements stringer
func (a AtomicSwap) String() string {
	return fmt.Sprintf("Atomic Swap"+
		"\n    ID:                       %s"+
		"\n    Status:                   %s"+
		"\n    Amount:                   %s"+
		"\n    Random number hash:       %s"+
		"\n    Expire height:            %d"+
		"\n    Timestamp:                %d"+
		"\n    Sender:                   %s"+
		"\n    Recipient:                %s"+
		"\n    Sender other chain:       %s"+
		"\n    Recipient other chain:    %s"+
		"\n    Closed block:             %d"+
		"\n    Cross chain:              %t"+
		"\n    Direction:                %s",
		a.GetSwapID(), a.Status.String(), a.Amount.String(),
		hex.EncodeToString(a.RandomNumberHash), a.ExpireHeight,
		a.Timestamp, a.Sender.String(), a.Recipient.String(),
		a.SenderOtherChain, a.RecipientOtherChain, a.ClosedBlock,
		a.CrossChain, a.Direction)
}

// AtomicSwaps is a slice of AtomicSwap
type AtomicSwaps []AtomicSwap

// String implements stringer
func (swaps AtomicSwaps) String() string {
	out := ""
	for _, swap := range swaps {
		out += swap.String() + "\n"
	}
	return out
}

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

// SwapDirection is the direction of an AtomicSwap
type SwapDirection byte

const (
	INVALID  SwapDirection = 0x00
	Incoming SwapDirection = 0x01
	Outgoing SwapDirection = 0x02
)

// NewSwapDirectionFromString converts string to SwapDirection type
func NewSwapDirectionFromString(str string) SwapDirection {
	switch str {
	case "Incoming", "incoming", "inc", "I", "i":
		return Incoming
	case "Outgoing", "outgoing", "out", "O", "o":
		return Outgoing
	default:
		return INVALID
	}
}

// String returns the string representation of a SwapDirection
func (direction SwapDirection) String() string {
	switch direction {
	case Incoming:
		return "Incoming"
	case Outgoing:
		return "Outgoing"
	default:
		return "INVALID"
	}
}

// MarshalJSON marshals the SwapDirection
func (direction SwapDirection) MarshalJSON() ([]byte, error) {
	return json.Marshal(direction.String())
}

// UnmarshalJSON unmarshals the SwapDirection
func (direction *SwapDirection) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*direction = NewSwapDirectionFromString(s)
	return nil
}
