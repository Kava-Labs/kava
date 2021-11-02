package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

// Permission is anything with a method that validates whether a proposal is allowed by it or not.
type Permission interface {
	Allows(sdk.Context, *codec.Codec, ParamKeeper, PubProposal) bool
}

func PackPermissions(permissions []Permission) []*types.Any {
	permissionsAny := make([]*types.Any, len(permissions))
	for i, permission := range permissions {
		msg, ok := permission.(proto.Message)
		if !ok {
			panic(fmt.Errorf("cannot proto marshal %T", permission))
		}
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
		permissionsAny[i] = any
	}
	return permissionsAny
}
