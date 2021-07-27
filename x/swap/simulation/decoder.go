package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/kava-labs/kava/x/swap/types"
)

// DecodeStore unmarshals the KVPair's Value to the module's corresponding type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.PoolKeyPrefix):
		var poolRecordA, poolRecordB types.PoolRecord
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &poolRecordA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &poolRecordB)
		return fmt.Sprintf("%v\n%v", poolRecordA, poolRecordB)
	case bytes.Equal(kvA.Key[:1], types.DepositorPoolSharesPrefix):
		var shareRecordA, shareRecordB types.ShareRecord
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &shareRecordA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &shareRecordB)
		return fmt.Sprintf("%v\n%v", shareRecordA, shareRecordB)
	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
