package types

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

type ParamKeeper interface {
	GetSubspace(string) (params.Subspace, bool)
}
