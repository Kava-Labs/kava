package txs

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
)

// SendTxPostPrice sends a tx containing MsgPostPrice to the kava blockchain
func SendTxPostPrice(
	chainID string,
	cdc *codec.Codec,
	accAddress sdk.AccAddress,
	moniker string,
	passphrase string,
	cliCtx context.CLIContext,
	msg *pftypes.MsgPostPrice,
	rpcURL string,
) (sdk.TxResponse, error) {

	if rpcURL != "" {
		cliCtx = cliCtx.WithNodeURI(rpcURL)
	}

	cliCtx.SkipConfirm = true

	txBldr := authtypes.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(cdc)).
		WithChainID(chainID)

	accountRetriever := authtypes.NewAccountRetriever(cliCtx)

	err := accountRetriever.EnsureExists(accAddress)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	// Prepare tx
	txBldr, err = utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	// Build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(moniker, passphrase, []sdk.Msg{msg})
	if err != nil {
		return sdk.TxResponse{}, err
	}

	// Broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTxCommit(txBytes)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	return res, nil
}

// ConstructMsgPostPrice builds a MsgPostPrice
func ConstructMsgPostPrice(accAddress sdk.AccAddress, price sdk.Dec, symbol string) (pftypes.MsgPostPrice, error) {
	// Set expiration time to 1 day in the future
	expiry := time.Now().Add(24 * time.Hour)

	// Initialize and validate the msg
	msg := pftypes.NewMsgPostPrice(accAddress, symbol, price, expiry)
	err := msg.ValidateBasic()
	if err != nil {
		return pftypes.MsgPostPrice{}, err
	}

	return msg, nil
}
