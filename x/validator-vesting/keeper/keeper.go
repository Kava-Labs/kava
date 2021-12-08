package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

// Keeper of the validatorvesting store
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.Codec
	bk       types.BankKeeper
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc codec.Codec, bk types.BankKeeper) Keeper {
	return Keeper{
		cdc: cdc,
		bk:  bk,
	}
}
