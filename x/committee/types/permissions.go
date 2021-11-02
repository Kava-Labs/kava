package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(sdk.Context, *codec.Codec, ParamKeeper, PubProposal) bool
}
