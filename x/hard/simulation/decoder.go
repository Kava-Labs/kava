package simulation

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/hard/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding hard type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.PreviousBlockTimeKey):
		var timeA, timeB time.Time
		cdc.MustUnmarshalBinaryBare(kvA.Value, &timeA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &timeB)
		return fmt.Sprintf("%s\n%s", timeA, timeB)

	case bytes.Equal(kvA.Key[:1], types.DepositsKeyPrefix):
		var depA, depB types.Deposit
		cdc.MustUnmarshalBinaryBare(kvA.Value, &depA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &depB)
		return fmt.Sprintf("%s\n%s", depA, depB)
	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
