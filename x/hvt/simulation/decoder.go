package simulation

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/hvt/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding cdp type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.PreviousBlockTimeKey):
		var timeA, timeB time.Time
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &timeA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &timeB)
		return fmt.Sprintf("%s\n%s", timeA, timeB)

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
