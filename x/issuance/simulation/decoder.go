package simulation

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/issuance/types"

	"github.com/tendermint/tendermint/libs/kv"
)

// DecodeStore the issuance module has no store keys -- all state is stored in params
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.AssetSupplyPrefix):
		var supplyA, supplyB types.AssetSupply
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
		return fmt.Sprintf("%s\n%s", supplyA, supplyB)
	case bytes.Equal(kvA.Key[:1], types.PreviousBlockTimeKey):
		var timeA, timeB time.Time
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &timeA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &timeB)
		return fmt.Sprintf("%s\n%s", timeA, timeB)
	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))

	}

}
