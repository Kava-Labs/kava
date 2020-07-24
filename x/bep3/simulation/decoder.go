package simulation

import (
	"bytes"
	"fmt"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/bep3/types"
)

// DecodeStore unmarshals the KVPair's Value to the module's corresponding type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.AtomicSwapKeyPrefix):
		var swapA, swapB types.AtomicSwap
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &swapA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &swapB)
		return fmt.Sprintf("%v\n%v", swapA, swapB)

	case bytes.Equal(kvA.Key[:1], types.AtomicSwapByBlockPrefix),
		bytes.Equal(kvA.Key[:1], types.AtomicSwapLongtermStoragePrefix):
		var bytesA tmbytes.HexBytes = kvA.Value
		var bytesB tmbytes.HexBytes = kvA.Value
		return fmt.Sprintf("%s\n%s", bytesA.String(), bytesB.String())

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
