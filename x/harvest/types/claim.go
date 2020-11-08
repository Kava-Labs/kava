package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Claim defines an amount of coins that the owner can claim
type Claim struct {
	Owner        sdk.AccAddress `json:"owner" yaml:"owner"`
	DepositDenom string         `json:"deposit_denom" yaml:"deposit_denom"`
	Amount       sdk.Coin       `json:"amount" yaml:"amount"`
	Type         ClaimType      `json:"type" yaml:"type"`
}

// NewClaim returns a new claim
func NewClaim(owner sdk.AccAddress, denom string, amount sdk.Coin, claimType ClaimType) Claim {
	return Claim{
		Owner:        owner,
		DepositDenom: denom,
		Amount:       amount,
		Type:         claimType,
	}
}
