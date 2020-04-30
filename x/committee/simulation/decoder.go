package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/libs/kv"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding module type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	// TODO implement this
	return ""
}
