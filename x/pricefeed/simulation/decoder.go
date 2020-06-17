package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	gogotypes "github.com/gogo/protobuf/types"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding pricefeed type
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB tmkv.Pair) string {
	return func(kvA, kvB tmkv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.CurrentPricePrefix):
			var priceA, priceB gogotypes.UInt64Value
			cdc.MustUnmarshalBinaryBare(kvA.Value, &priceA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &priceB)
			return fmt.Sprintf("%s\n%s", priceA, priceB)

		case bytes.Contains(kvA.Key, []byte(types.RawPriceFeedPrefix)):
			var postedPriceA, postedPriceB []types.PostedPrice
			cdc.MustUnmarshalBinaryBare(kvA.Value, &postedPriceA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &postedPriceB)
			return fmt.Sprintf("%s\n%s", postedPriceA, postedPriceB)

		default:
			panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
		}
	}
}
