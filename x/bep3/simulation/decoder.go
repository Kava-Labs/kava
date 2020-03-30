package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// DecodeStore unmarshals the KVPair's Value to the module's corresponding type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	// TODO implement this
	return ""
}
