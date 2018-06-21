package lcd

import (
	"net/http"
	//"encoding/hex"
	"encoding/base64"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

type TxBody struct {
	TxBase64 string `json:"txbase64"`
}

// Decode a tx from base64 into json
func DecodeTxRequestHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get the input base64 string
		var m TxBody
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&m)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}

		// convert from base64 string to bytes
		txBytes, err := base64.StdEncoding.DecodeString(m.TxBase64)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}

		// convert bytes to Tx struct
		var tx auth.StdTx
		err = cdc.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}

		// convert Tx struct to json (bytes) and return
		output, err := cdc.MarshalJSON(tx)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
