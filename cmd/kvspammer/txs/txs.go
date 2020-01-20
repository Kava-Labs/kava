package txs

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
)

// Gas is set to a large amount to ensure successful tx sending
const Gas = 500000

// SendTxRPC sends a tx containing a sdk.Msg to the kava blockchain
func SendTxRPC(
	chainID string,
	cdc *codec.Codec,
	accAddress sdk.AccAddress,
	moniker string,
	passphrase string,
	cliCtx context.CLIContext,
	msg []sdk.Msg,
	rpcURL string,
) (sdk.TxResponse, error) {

	if rpcURL != "" {
		cliCtx = cliCtx.WithNodeURI(rpcURL)
	}

	cliCtx.SkipConfirm = true

	txBldr := authtypes.NewTxBuilderFromCLI().
		WithTxEncoder(utils.GetTxEncoder(cdc)).
		WithChainID(chainID).
		WithGas(Gas)

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
	txBytes, err := txBldr.BuildAndSign(moniker, passphrase, msg)
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

// QueryCDP queries an individual CDP
func QueryCDP(
	cdc *codec.Codec,
	cliCtx context.CLIContext,
	accAddress sdk.AccAddress,
	collateralDenom string,
) (cdptypes.CDP, bool, error) {

	bz, err := cdc.MarshalJSON(cdptypes.QueryCdpParams{
		CollateralDenom: collateralDenom,
		Owner:           accAddress,
	})
	if err != nil {
		return cdptypes.CDP{}, false, err
	}

	// Query
	route := fmt.Sprintf("custom/cdp/cdp")
	res, _, err := cliCtx.QueryWithData(route, bz)
	if err != nil {
		return cdptypes.CDP{}, false, err
	}

	// Decode and print results
	var cdp cdptypes.CDP
	cdc.MustUnmarshalJSON(res, &cdp)
	return cdp, true, nil
}
