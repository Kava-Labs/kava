package v032

import (
	"github.com/cosmos/cosmos-sdk/codec"
	//amino "github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/tendermint/tendermint/types"
)

// Replace amino codec with sdk codec to avoid an explicit amino import in go.mod.
// This will use a different version of amino from tendermint v0.32, but they are backwards compatible.
var Cdc = codec.New()

func init() {
	RegisterBlockAmino(Cdc)
}

func RegisterBlockAmino(cdc *codec.Codec) {
	cryptoAmino.RegisterAmino(cdc)
	types.RegisterEvidences(cdc) // v0.33 is backwards compatible with v0.32 here
}
