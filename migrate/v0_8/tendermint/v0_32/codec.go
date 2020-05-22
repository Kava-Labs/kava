package v032

import (
	amino "github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/tendermint/tendermint/types"
)

var Cdc = amino.NewCodec()

func init() {
	RegisterBlockAmino(Cdc)
}

func RegisterBlockAmino(cdc *amino.Codec) {
	cryptoAmino.RegisterAmino(cdc)
	types.RegisterEvidences(cdc) // v0.33 is backwards compatible with v0.32 here
}

// // GetCodec returns a codec used by the package. For testing purposes only.
// func GetCodec() *amino.Codec {
// 	return cdc
// }

// // For testing purposes only
// func RegisterMockEvidencesGlobal() {
// 	RegisterMockEvidences(cdc)
// }
