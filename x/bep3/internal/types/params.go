package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params"
	// "github.com/kava-labs/kava/x/bep3/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Parameter keys
var (
	KeyMinimumLockTime = []byte("MinimumLockTime")
	// TODO: validate this time as reasonable
	DefaultMinimumLockTime time.Duration = 1 * time.Hour
)

// Params governance parameters for bep3 module
type Params struct {
	MinimumLockTime time.Duration `json:"minimum_lock_time" yaml:"minimum_lock_time"`
}

// // GetParams returns the params from the store
// func (k Keeper) GetParams(ctx sdk.Context) types.Params {
// 	var p types.Params
// 	k.paramSubspace.GetParamSet(ctx, &p)
// 	return p
// }

// // SetParams sets params on the store
// func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
// 	k.paramSubspace.SetParamSet(ctx, &params)
// }

// NewParams returns a new params object
func NewParams(minimumLockTime time.Duration) Params {
	return Params{
		MinimumLockTime: minimumLockTime,
	}
}

// DefaultParams returns default params for bep3 module
func DefaultParams() Params {
	return NewParams(DefaultMinimumLockTime)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() subspace.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{Key: KeyMinimumLockTime, Value: &p.MinimumLockTime},
	}
}

// String implements fmt.Stringer
func (p Params) String() string {
	return fmt.Sprintf("Params: \n Minimum Lock Time: %s", p.MinimumLockTime)
}

// Validate ensure that params have valid values
func (p Params) Validate() error {
	if p.MinimumLockTime <= 0 {
		return sdk.ErrInternal("minimum lock time must be greater than 0")
	}
	return nil
}
