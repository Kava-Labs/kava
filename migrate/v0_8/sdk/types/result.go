package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxResponse defines a structure containing relevant tx data and metadata. The
// tags are stringified and the log is JSON decoded.
type TxResponse struct {
	Height    int64            `json:"height"`
	TxHash    string           `json:"txhash"`
	Code      uint32           `json:"code,omitempty"`
	Data      string           `json:"data,omitempty"`
	RawLog    string           `json:"raw_log,omitempty"`
	Logs      ABCIMessageLogs  `json:"logs,omitempty"`
	Info      string           `json:"info,omitempty"`
	GasWanted int64            `json:"gas_wanted,omitempty"`
	GasUsed   int64            `json:"gas_used,omitempty"`
	Codespace string           `json:"codespace,omitempty"`
	Tx        sdk.Tx           `json:"tx,omitempty"`
	Timestamp string           `json:"timestamp,omitempty"`
	Events    sdk.StringEvents `json:"events,omitempty"`
}

// Empty returns true if the response is empty
func (r TxResponse) Empty() bool {
	return r.TxHash == "" && r.Logs == nil
}

// ABCIMessageLogs represents a slice of ABCIMessageLog.
type ABCIMessageLogs []ABCIMessageLog

// ABCIMessageLog defines a structure containing an indexed tx ABCI message log.
type ABCIMessageLog struct {
	MsgIndex uint16 `json:"msg_index"`
	Success  bool   `json:"success"`
	Log      string `json:"log"`

	// Events contains a slice of Event objects that were emitted during some
	// execution.
	Events sdk.StringEvents `json:"events"`
}
