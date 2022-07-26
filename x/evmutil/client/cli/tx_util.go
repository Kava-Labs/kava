package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// GenerateOrBroadcastTx checks CLI flags and generates or broadcasts a
// transaction this is used over tx.GenerateOrBroadcastTxCLI as it does not sign
// the message.
func GenerateOrBroadcastTx(clientCtx client.Context, signingTx signing.Tx) error {
	// These manual flag checks are required as we use broadcast the tx
	// directly via BroadcastTx instead of tx.GenerateOrBroadcastTxCLI
	// which handles flags for us.

	if clientCtx.GenerateOnly {
		if err := PrintTx(clientCtx, signingTx); err != nil {
			return err
		}
	}

	if err := ConfirmTx(clientCtx, signingTx); err != nil {
		return err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(signingTx)
	if err != nil {
		return err
	}

	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

// PrintTx outputs a signing.Tx in JSON format, ie. when the GenerateOnly flag
// is enabled.
func PrintTx(clientCtx client.Context, signingTx signing.Tx) error {
	json, err := clientCtx.TxConfig.TxJSONEncoder()(signingTx)
	if err != nil {
		return err
	}

	return clientCtx.PrintString(fmt.Sprintf("%s\n", json))
}

// ConfirmTx outputs the transaction to be signed and requests confirmation
// if the SkipConfirm flag is not enabled.
func ConfirmTx(clientCtx client.Context, signingTx signing.Tx) error {
	if clientCtx.SkipConfirm {
		return nil
	}

	out, err := clientCtx.TxConfig.TxJSONEncoder()(signingTx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n\n", out)
	if err != nil {
		return err
	}

	buf := bufio.NewReader(os.Stdin)
	ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)
	if err != nil {
		return err
	}

	if !ok {
		_, err = fmt.Fprintf(os.Stderr, "%s\n", "cancelled transaction")
		if err != nil {
			return err
		}
	}

	return nil
}
