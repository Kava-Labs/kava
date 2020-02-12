package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SwapStatus byte

const (
	NULL      SwapStatus = 0x00
	Open      SwapStatus = 0x01
	Completed SwapStatus = 0x02
	Expired   SwapStatus = 0x03
)

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

func (status SwapStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

func (status *SwapStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*status = NewSwapStatusFromString(s)
	return nil
}

type SwapBytes []byte

func (bz SwapBytes) Marshal() ([]byte, error) {
	return bz, nil
}

func (bz *SwapBytes) Unmarshal(data []byte) error {
	*bz = data
	return nil
}

func (bz SwapBytes) MarshalJSON() ([]byte, error) {
	s := hex.EncodeToString(bz)
	jbz := make([]byte, len(s)+2)
	jbz[0] = '"'
	copy(jbz[1:], []byte(s))
	jbz[len(jbz)-1] = '"'
	return jbz, nil
}

func (bz *SwapBytes) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("Invalid hex string: %s", data)
	}
	bz2, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*bz = bz2
	return nil
}

type AtomicSwap struct {
	From      sdk.AccAddress `json:"from"`
	To        sdk.AccAddress `json:"to"`
	OutAmount sdk.Coins      `json:"out_amount"`
	InAmount  sdk.Coins      `json:"in_amount"`

	ExpectedIncome      string `json:"expected_income"`
	RecipientOtherChain string `json:"recipient_other_chain"`

	RandomNumberHash SwapBytes `json:"random_number_hash"`
	RandomNumber     SwapBytes `json:"random_number"`
	Timestamp        int64     `json:"timestamp"`

	CrossChain bool `json:"cross_chain"`

	ExpireHeight int64      `json:"expire_height"`
	Index        int64      `json:"index"`
	ClosedTime   int64      `json:"closed_time"`
	Status       SwapStatus `json:"status"`
}

// Params for query 'custom/bep3/swapid'
type QuerySwapByID struct {
	SwapID SwapBytes
}

// Params for query 'custom/bep3/swapcreator'
type QuerySwapByCreatorParams struct {
	Creator sdk.AccAddress
	Limit   int64
	Offset  int64
}

// Params for query 'custom/bep3/swaprecipient'
type QuerySwapByRecipientParams struct {
	Recipient sdk.AccAddress
	Limit     int64
	Offset    int64
}
