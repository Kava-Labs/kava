package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/kava-labs/kava/x/bep3/types"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding bep3 type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.AtomicSwapKeyPrefix):
		var swapA, swapB *types.AtomicSwap
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &swapA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &swapB)
		return fmt.Sprintf("%v\n%v", swapA, swapB)

	case bytes.Equal(kvA.Key[:1], types.AssetSupplyKeyPrefix):
		var supplyA, supplyB types.AssetSupply
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
		return fmt.Sprintf("%s\n%s", supplyA, supplyB)

	case bytes.Equal(kvA.Key[:1], types.AtomicSwapByBlockPrefix),
		bytes.Equal(kvA.Key[:1], types.AtomicSwapLongtermStoragePrefix):
		var bytesA cmn.HexBytes = kvA.Value
		var bytesB cmn.HexBytes = kvA.Value
		return fmt.Sprintf("%s\n%s", bytesA.String(), bytesB.String())

	case bytes.Equal(kvA.Key[:1], types.AtomicSwapByBlockPrefix),
		bytes.Equal(kvA.Key[:1], types.AtomicSwapLongtermStoragePrefix):
		var bytesA cmn.HexBytes = kvA.Value
		var bytesB cmn.HexBytes = kvA.Value
		return fmt.Sprintf("%s\n%s", bytesA.String(), bytesB.String())

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
