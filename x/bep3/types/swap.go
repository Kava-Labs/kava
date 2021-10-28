package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewAtomicSwap returns a new AtomicSwap
func NewAtomicSwap(amount sdk.Coins, randomNumberHash tmbytes.HexBytes, expireHeight uint64, timestamp int64,
	sender, recipient string, senderOtherChain string, recipientOtherChain string, closedBlock int64,
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

// Validate performs a basic validation of an atomic swap fields.
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
	if len(a.Sender) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender cannot be empty")
	}
	if len(a.Recipient) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient cannot be empty")
	}
	if len(a.Sender) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Sender))
	}
	if len(a.Recipient) != AddrByteCount {
		return fmt.Errorf("the expected address length is %d, actual length is %d", AddrByteCount, len(a.Recipient))
	}
	// NOTE: These adresses may not have a bech32 prefix.
	if strings.TrimSpace(a.SenderOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender other chain cannot be blank")
	}
	if strings.TrimSpace(a.RecipientOtherChain) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "recipient other chain cannot be blank")
	}
	if a.Status == SWAP_STATUS_COMPLETED && a.ClosedBlock == 0 {
		return errors.New("closed block cannot be 0")
	}
	if a.Status == SWAP_STATUS_UNSPECIFIED || a.Status > 3 {
		return errors.New("invalid swap status")
	}
	if a.Direction == SWAP_DIRECTION_UNSPECIFIED || a.Direction > 2 {
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
		a.Timestamp, a.Sender, a.Recipient,
		a.SenderOtherChain, a.RecipientOtherChain, a.ClosedBlock,
		a.CrossChain, a.Direction)
}

// IsValid returns true if the swap status is valid and false otherwise.
func (status SwapStatus) IsValid() bool {
	if status == SWAP_STATUS_OPEN ||
		status == SWAP_STATUS_COMPLETED ||
		status == SWAP_STATUS_EXPIRED {
		return true
	}
	return false
}

// IsValid returns true if the swap direction is valid and false otherwise.
func (direction SwapDirection) IsValid() bool {
	if direction == SWAP_DIRECTION_INCOMING ||
		direction == SWAP_DIRECTION_OUTGOING {
		return true
	}
	return false
}

func NewAugmentedAtomicSwap(swap AtomicSwap) AugmentedAtomicSwap {
	return AugmentedAtomicSwap{
		ID:         hex.EncodeToString(swap.GetSwapID()),
		AtomicSwap: swap,
	}
}

func (a *AugmentedAtomicSwap) String() string {
	return fmt.Sprintf("Augmented Atomic Swap"+
		"\n    ID:     %s"+
		"\n%s",
		a.ID,
		a.AtomicSwap.String())
}

// LegacyAugmentedAtomicSwap defines an ID and AtomicSwap fields on the top level.
// This should be removed when legacy REST endpoints are removed.
type LegacyAugmentedAtomicSwap struct {
	ID string `json:"id" yaml:"id"`

	// Embed AtomicSwap fields explicity in order to output as top level JSON fields
	// This prevents breaking changes for clients using REST API
	Amount              sdk.Coins        `json:"amount"  yaml:"amount"`
	RandomNumberHash    tmbytes.HexBytes `json:"random_number_hash"  yaml:"random_number_hash"`
	ExpireHeight        uint64           `json:"expire_height"  yaml:"expire_height"`
	Timestamp           int64            `json:"timestamp"  yaml:"timestamp"`
	Sender              string           `json:"sender"  yaml:"sender"`
	Recipient           string           `json:"recipient"  yaml:"recipient"`
	SenderOtherChain    string           `json:"sender_other_chain"  yaml:"sender_other_chain"`
	RecipientOtherChain string           `json:"recipient_other_chain"  yaml:"recipient_other_chain"`
	ClosedBlock         int64            `json:"closed_block"  yaml:"closed_block"`
	Status              SwapStatus       `json:"status"  yaml:"status"`
	CrossChain          bool             `json:"cross_chain"  yaml:"cross_chain"`
	Direction           SwapDirection    `json:"direction"  yaml:"direction"`
}

func NewLegacyAugmentedAtomicSwap(swap AtomicSwap) LegacyAugmentedAtomicSwap {
	return LegacyAugmentedAtomicSwap{
		ID:                  hex.EncodeToString(swap.GetSwapID()),
		Amount:              swap.Amount,
		RandomNumberHash:    swap.RandomNumberHash,
		ExpireHeight:        swap.ExpireHeight,
		Timestamp:           swap.Timestamp,
		Sender:              swap.Sender,
		Recipient:           swap.Recipient,
		SenderOtherChain:    swap.SenderOtherChain,
		RecipientOtherChain: swap.RecipientOtherChain,
		ClosedBlock:         swap.ClosedBlock,
		Status:              swap.Status,
		CrossChain:          swap.CrossChain,
		Direction:           swap.Direction,
	}
}
