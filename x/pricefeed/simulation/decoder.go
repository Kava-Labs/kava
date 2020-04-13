package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding pricefeed type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Contains(kvA.Key, []byte(types.CurrentPricePrefix)):
		var priceA, priceB types.CurrentPrice
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
