package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/x/issuance/types"

	"github.com/tendermint/tendermint/libs/kv"
)

// DecodeStore the issuance module has no store keys -- all state is stored in params
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
}
