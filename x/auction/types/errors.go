package types

import errorsmod "cosmossdk.io/errors"

// DONTCOVER

var (
	// ErrInvalidInitialAuctionID error for when the initial auction ID hasn't been set
	ErrInvalidInitialAuctionID = errorsmod.Register(ModuleName, 2, "initial auction ID hasn't been set")
	// ErrUnrecognizedAuctionType error for unrecognized auction type
	ErrUnrecognizedAuctionType = errorsmod.Register(ModuleName, 3, "unrecognized auction type")
	// ErrAuctionNotFound error for when an auction is not found
	ErrAuctionNotFound = errorsmod.Register(ModuleName, 4, "auction not found")
	// ErrAuctionHasNotExpired error for attempting to close an auction that has not passed its end time
	ErrAuctionHasNotExpired = errorsmod.Register(ModuleName, 5, "auction can't be closed as curent block time has not passed auction end time")
	// ErrAuctionHasExpired error for when an auction is closed and unavailable for bidding
	ErrAuctionHasExpired = errorsmod.Register(ModuleName, 6, "auction has closed")
	// ErrInvalidBidDenom error for when bid denom doesn't match auction bid denom
	ErrInvalidBidDenom = errorsmod.Register(ModuleName, 7, "bid denom doesn't match auction bid denom")
	// ErrInvalidLotDenom error for when lot denom doesn't match auction lot denom
	ErrInvalidLotDenom = errorsmod.Register(ModuleName, 8, "lot denom doesn't match auction lot denom")
	// ErrBidTooSmall error for when bid is not greater than auction's min bid amount
	ErrBidTooSmall = errorsmod.Register(ModuleName, 9, "bid is not greater than auction's min new bid amount")
	// ErrBidTooLarge error for when bid is larger than auction's maximum allowed bid
	ErrBidTooLarge = errorsmod.Register(ModuleName, 10, "bid is greater than auction's max bid")
	// ErrLotTooSmall error for when lot is less than zero
	ErrLotTooSmall = errorsmod.Register(ModuleName, 11, "lot is not greater than auction's min new lot amount")
	// ErrLotTooLarge error for when lot is not smaller than auction's max new lot amount
	ErrLotTooLarge = errorsmod.Register(ModuleName, 12, "lot is greater than auction's max new lot amount")
)
