package simulation

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding auth type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.ValidatorVestingAccountPrefix):
		var accA, accB sdk.AccAddress
		accA = sdk.AccAddress(kvA.Key[1:])
		accB = sdk.AccAddress(kvB.Key[1:])
		return fmt.Sprintf("%v\n%v", accA, accB)
	case bytes.Equal(kvA.Key, types.BlocktimeKey):
		var btA, btB time.Time
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &btA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &btB)
		return fmt.Sprintf("%v\n%v", btA, btB)
	default:
		panic(fmt.Sprintf("invalid %s key %X", types.ModuleName, kvA.Key))
	}
}
