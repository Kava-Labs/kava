package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SeizedDebt tracks debt seized from liquidated CDPs.
type SeizedDebt struct {
	Total         sdk.Int // Total debt seized from CDPs. Known as Awe in maker.
	SentToAuction sdk.Int // Portion of seized debt that has had a (reverse) auction was started for it. Known as Ash in maker.
	// SentToAuction should always be < Total
}

// Available gets the seized debt that has not been sent for auction. Known as Woe in maker.
func (sd SeizedDebt) Available() sdk.Int {
	return sd.Total.Sub(sd.SentToAuction)
}

// Settle reduces the amount of debt
func (sd SeizedDebt) Settle(amount sdk.Int) (SeizedDebt, sdk.Error) {
	if amount.IsNegative() {
		return sd, sdk.ErrInternal("tried to settle a negative amount")
	}
	if amount.GT(sd.Total) {
		return sd, sdk.ErrInternal("tried to settle more debt than exists")
	}
	sd.Total = sd.Total.Sub(amount)
	sd.SentToAuction = sdk.MaxInt(sd.SentToAuction.Sub(amount), sdk.ZeroInt())
	return sd, nil
}
