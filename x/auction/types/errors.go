// DONTCOVER
package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Error codes specific to auction module
const (
	DefaultCodespace                      sdk.CodespaceType = ModuleName
	CodeInvalidInitialAuctionID           sdk.CodeType      = 1
	CodeInvalidModulePermissions          sdk.CodeType      = 2
	CodeUnrecognizedAuctionType           sdk.CodeType      = 3
	CodeAuctionNotFound                   sdk.CodeType      = 4
	CodeAuctionHasNotExpired              sdk.CodeType      = 5
	CodeAuctionHasExpired                 sdk.CodeType      = 6
	CodeInvalidBidDenom                   sdk.CodeType      = 7
	CodeInvalidLotDenom                   sdk.CodeType      = 8
	CodeBidTooSmall                       sdk.CodeType      = 9
	CodeBidTooLarge                       sdk.CodeType      = 10
	CodeLotTooLarge                       sdk.CodeType      = 11
	CodeCollateralAuctionIsInReversePhase sdk.CodeType      = 12
	CodeCollateralAuctionIsInForwardPhase sdk.CodeType      = 13
)

// ErrInvalidInitialAuctionID error for when the initial auction ID hasn't been set
func ErrInvalidInitialAuctionID(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInitialAuctionID, fmt.Sprintf("initial auction ID hasn't been set"))
}

// ErrInvalidModulePermissions error for when module doesn't have valid permissions
func ErrInvalidModulePermissions(codespace sdk.CodespaceType, permission string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidModulePermissions, fmt.Sprintf("module does not have required permission '%s'", permission))
}

// ErrUnrecognizedAuctionType error for unrecognized auction type
func ErrUnrecognizedAuctionType(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeUnrecognizedAuctionType, fmt.Sprintf("unrecognized auction type"))
}

// ErrAuctionNotFound error for when an auction is not found
func ErrAuctionNotFound(codespace sdk.CodespaceType, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeAuctionNotFound, fmt.Sprintf("auction %d was not found", id))
}

// ErrAuctionHasNotExpired error for attempting to close an auction that has not passed its end time
func ErrAuctionHasNotExpired(codespace sdk.CodespaceType, blockTime time.Time, endTime time.Time) sdk.Error {
	return sdk.NewError(codespace, CodeAuctionHasNotExpired, fmt.Sprintf("auction can't be closed as curent block time (%v) has not passed auction end time (%v)", blockTime, endTime))
}

// ErrAuctionHasExpired error for when an auction is closed and unavailable for bidding
func ErrAuctionHasExpired(codespace sdk.CodespaceType, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeAuctionHasExpired, fmt.Sprintf("auction %d has closed", id))
}

// ErrInvalidBidDenom error for when bid denom doesn't match auction bid denom
func ErrInvalidBidDenom(codespace sdk.CodespaceType, bidDenom string, auctionBidDenom string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBidDenom, fmt.Sprintf("bid denom %s doesn't match auction bid denom %s", bidDenom, auctionBidDenom))
}

// ErrInvalidLotDenom error for when lot denom doesn't match auction lot denom
func ErrInvalidLotDenom(codespace sdk.CodespaceType, lotDenom string, auctionLotDenom string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidLotDenom, fmt.Sprintf("lot denom %s doesn't match auction lot denom %s", lotDenom, auctionLotDenom))
}

// ErrBidTooSmall error for when bid is not greater than auction's last bid
func ErrBidTooSmall(codespace sdk.CodespaceType, bid sdk.Coin, lastBid sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeBidTooSmall, fmt.Sprintf("bid %s is not greater than auction's last bid %s", bid.String(), lastBid.String()))
}

// ErrBidTooLarge error for when bid is larger than auction's maximum allowed bid
func ErrBidTooLarge(codespace sdk.CodespaceType, bid sdk.Coin, maxBid sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeBidTooLarge, fmt.Sprintf("bid %s is greater than auction's max bid %s", bid.String(), maxBid.String()))
}

// ErrLotTooLarge error for when lot is not smaller than auction's last lot
func ErrLotTooLarge(codespace sdk.CodespaceType, lot sdk.Coin, lastLot sdk.Coin) sdk.Error {
	return sdk.NewError(codespace, CodeLotTooLarge, fmt.Sprintf("lot %s is not less than auction's last lot %s", lot.String(), lastLot.String()))
}

// ErrCollateralAuctionIsInReversePhase error for when attempting to place a forward bid on a collateral auction in reverse phase
func ErrCollateralAuctionIsInReversePhase(codespace sdk.CodespaceType, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeCollateralAuctionIsInReversePhase, fmt.Sprintf("invalid bid - auction %d is in reverse phase", id))
}

// ErrCollateralAuctionIsInForwardPhase error for when attempting to place a reverse bid on a collateral auction in forward phase
func ErrCollateralAuctionIsInForwardPhase(codespace sdk.CodespaceType, id uint64) sdk.Error {
	return sdk.NewError(codespace, CodeCollateralAuctionIsInForwardPhase, fmt.Sprintf("invalid bid - auction %d is in forward phase", id))
}
