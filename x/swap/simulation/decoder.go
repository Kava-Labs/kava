package simulation

import (
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DecodeStore unmarshals the KVPair's Value to the module's corresponding type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	// TODO: as store keys are added to the module, test marshal/unmarshal of each key prefix
	return ""
}
