package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeProposalExpired  sdk.CodeType = 1
	CodeUnknownItem      sdk.CodeType = 2
	CodeInvalidGenesis   sdk.CodeType = 3
	CodeInvalidProposal  sdk.CodeType = 4
	CodeInvalidCommittee sdk.CodeType = 5
)

func ErrUnknownCommittee(codespace sdk.CodespaceType, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownItem, fmt.Sprintf("committee with id '%d' not found", id))
}

func ErrInvalidCommittee(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidCommittee, msg)
}

func ErrUnknownProposal(codespace sdk.CodespaceType, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownItem, fmt.Sprintf("proposal with id '%d' not found", id))
}

func ErrProposalExpired(codespace sdk.CodespaceType, blockTime, expiry time.Time) sdk.Error {
	return sdk.NewError(codespace, CodeProposalExpired, fmt.Sprintf("proposal expired at %s, current blocktime %s", expiry, blockTime))
}

func ErrInvalidPubProposal(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProposal, msg)
}

func ErrUnknownVote(codespace sdk.CodespaceType, proposalID uint64, voter sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownItem, fmt.Sprintf("vote with for proposal '%d' and voter %s not found", proposalID, voter))
}

func ErrInvalidGenesis(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidGenesis, msg)
}

func ErrNoProposalHandlerExists(codespace sdk.CodespaceType, content interface{}) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownItem, fmt.Sprintf("'%T' does not have a corresponding handler", content))
}
