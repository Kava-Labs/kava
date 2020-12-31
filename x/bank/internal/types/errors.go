package types

import (
	cosmosbank "github.com/cosmos/cosmos-sdk/x/bank"
)

// x/bank module sentinel errors
var (
	ErrNoInputs            = cosmosbank.ErrNoInputs
	ErrNoOutputs           = cosmosbank.ErrNoOutputs
	ErrInputOutputMismatch = cosmosbank.ErrInputOutputMismatch
	ErrSendDisabled        = cosmosbank.ErrSendDisabled
)
