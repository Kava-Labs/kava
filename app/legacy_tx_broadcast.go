package app

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/gorilla/mux"
)

// RegisterLegacyTxRoutes registers a legacy tx routes that use amino encoding json
func RegisterLegacyTxRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/txs", legacyTxBroadcast(clientCtx)).Methods("POST")
}

// LegacyTxBroadcastRequest represents a broadcast request with an amino json encoded transaction
type LegacyTxBroadcastRequest struct {
	Tx   legacytx.StdTx `json:"tx"`
	Mode string         `json:"mode"`
}

var _ codectypes.UnpackInterfacesMessage = LegacyTxBroadcastRequest{}

// UnpackInterfaces implements the UnpackInterfacesMessage interface
func (m LegacyTxBroadcastRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Tx.UnpackInterfaces(unpacker)
}

func legacyTxBroadcast(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LegacyTxBroadcastRequest
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
			return
		}

		tx := req.Tx
		builder := clientCtx.TxConfig.NewTxBuilder()

		err := builder.SetMsgs(tx.GetMsgs()...)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		builder.SetFeeAmount(tx.GetFee())
		builder.SetGasLimit(tx.GetGas())
		builder.SetMemo(req.Tx.GetMemo())
		builder.SetTimeoutHeight(req.Tx.GetTimeoutHeight())

		signatures, err := tx.GetSignaturesV2()
		if rest.CheckBadRequestError(w, err) {
			return
		}
		for i, sig := range signatures {
			addr := sdk.AccAddress(sig.PubKey.Address())
			_, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, addr)
			if rest.CheckBadRequestError(w, err) {
				return
			}

			signatures[i].Sequence = seq
		}

		err = builder.SetSignatures(signatures...)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		txBytes, err := clientCtx.TxConfig.TxEncoder()(builder.GetTx())
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithBroadcastMode(req.Mode)
		res, err := clientCtx.BroadcastTx(txBytes)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponseBare(w, clientCtx, res)
	}
}
