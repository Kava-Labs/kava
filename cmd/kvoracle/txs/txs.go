package txs

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// SendTxPostPrice sends a tx containing MsgPostPrice to the kava blockchain
func SendTxPostPrice(
	chainID string,
	cdc *codec.Codec,
	accAddress sdk.AccAddress,
	moniker string,
	passphrase string,
	cliCtx context.CLIContext,
	msg *types.MsgPostPrice,
	rpcURL string,
) error {

	// if rpcURL != "" {
	// 	cliCtx = cliCtx.WithNodeURI(rpcURL)
	// }

	// cliCtx.SkipConfirm = true

	// txBldr := authtypes.NewTxBuilderFromCLI().
	// 	WithTxEncoder(utils.GetTxEncoder(cdc)).
	// 	WithChainID(chainID)

	// accountRetriever := authtypes.NewAccountRetriever(cliCtx)

	// acc, err := authtypes.NewAccountRetriever(cliCtx).GetAccount(accAddress)
	// if err != nil {
	// 	return err
	// }

	// // err := accountRetriever.EnsureExists((sdk.AccAddress(msg.From)))
	// // if err != nil {
	// // 	return err
	// // }

	// gasPrices := sdk.Coins{sdk.NewCoin("stake", sdk.NewInt(50))}
	// // gas := sdk.NewInt(200000)
	// gas := uint64(200000)

	// fees := make(sdk.Coins, len(gasPrices))
	// for i, gp := range gasPrices {
	// 	fee := gp.Amount.Mul(sdk.NewInt(int64(gas)))
	// 	fees[i] = sdk.NewCoin(gp.Denom, fee) //(fee).Ceil().RoundInt())
	// }

	// // Build the StdSignMsg
	// sign := authtypes.StdSignMsg{
	// 	ChainID:       chainID,
	// 	AccountNumber: acc.GetSequence(),
	// 	Sequence:      acc.GetSequence(),
	// 	Memo:          "",
	// 	Msgs:          msgs,
	// 	Fee:           authtypes.NewStdFee(gas, fees),
	// }

	// // Create signature for transaction
	// stdSignature, err := authtypes.MakeSignature(nil, moniker, passphrase, sign)

	// // Create the StdTx for broadcast
	// stdTx := authtypes.NewStdTx(msgs, sign.Fee, []authtypes.StdSignature{stdSignature}, "")
	// // Marshal amino
	// out, err := cdc.MarshalBinaryLengthPrefixed(stdTx)
	// if err != nil {
	// 	return err
	// }

	// Broadcast transaction
	// res, err := cliCtx.BroadcastTxSync(out) // BroadcastTxCommit
	// if err != nil {
	// 	return err
	// }
	// // Prepare tx
	// txBldr, err = utils.PrepareTxBuilder(txBldr, cliCtx)
	// if err != nil {
	// 	return err
	// }

	// // Build and sign the transaction
	// txBytes, err := txBldr.BuildAndSign(moniker, passphrase, []sdk.Msg{msg})
	// if err != nil {
	// 	return err
	// }

	// // Broadcast to a Tendermint node
	// res, err := cliCtx.BroadcastTxSync(txBytes)
	// if err != nil {
	// 	return err
	// }

	// cliCtx.PrintOutput(res)
	return nil
}
