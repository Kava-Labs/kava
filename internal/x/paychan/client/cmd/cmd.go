package cli

import (
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"

	"github.com/kava-labs/kava/internal/x/paychan"
)

// list of functions that return pointers to cobra commands
// No local storage needed for cli acting as a sender

// Currently minimum set of cli commands are implemented:
// create paychan - create and fund
// generate new paychan state - print a half signed close tx (sender signs)
// close paychan - close using state (receiver signs)

// Future cli commands:
// create paychan
// close paychan
// get paychan(s)
// send paychan payment
// get balance from receiver

/*
func CreatePaychanCmd(cdc *wire.Codec) *cobra.Command {
	flagTo := "to"
	flagAmount := "amount"

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new payment channel",
		Long:  "Create a new payment channel from a local address to a remote address, funded with some amount of coins. These coins are removed from the sender account and put into the payment channel.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a "client context" stuct populated with info from common flags
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// Get sender adress
			senderAddress, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// Get receiver address
			toStr := viper.GetString(flagTo)
			receiverAddress, err := sdk.GetAccAddressBech32(toStr)
			if err != nil {
				return err
			}

			// Get channel funding amount
			amountString := viper.GetString(flagAmount)
			amount, err := sdk.ParseCoins(amountString)
			if err != nil {
				return err
			}

			// Create the create channel msg to send
			// TODO write NewMsgCreate func?
			msg := paychan.MsgCreate{
				Sender:   senderAddress,
				Receiver: receiverAddress,
				Amount:   amount,
			}
			// Build and sign the transaction, then broadcast to the blockchain
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
	cmd.Flags().String(flagTo, "", "Recipient address of the payment channel")
	cmd.Flags().String(flagAmount, "", "Amount of coins to fund the paymetn channel with")
	return cmd
}

func GenerateNewStateCmd(cdc *wire.Codec) *cobra.Command {
	flagId := "id"
	flagTo := "to"
	flagAmount := "amount"

	cmd := &cobra.Command{
		Use:   "new-state",
		Short: "Generate a new payment channel state.",
		Long:  "Generate a new state for an existing payment channel and print it out. The new state is represented as a half signed close transaction, signed by the sender.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a "client context" stuct populated with info from common flags
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// Get sender adress
			senderAddress, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// Get the paychan id
			id := viper.GetInt64(flagId)

			// Get receiver address
			toStr := viper.GetString(flagTo)
			receiverAddress, err := sdk.GetAccAddressBech32(toStr)
			if err != nil {
				return err
			}

			// Get channel receiver amount
			amountString := viper.GetString(flagAmount)
			amount, err := sdk.ParseCoins(amountString)
			if err != nil {
				return err
			}

			// create close paychan msg
			msg := paychan.MsgClose{
				Sender:         senderAddress,
				Receiver:       receiverAddress,
				Id:             id,
				ReceiverAmount: amount,
			}

			// Sign the msg as the sender
			txBytes, err := EnsureSignBuild(ctx, ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			// Print out the signed msg
			fmt.Println("txBytes:", txBytes)
			//encodedTxBytes := make([]byte, base64.StdEncoding.EncodedLen(len(txBytes)))
			encodedTxBytes := base64.StdEncoding.EncodeToString(txBytes)
			fmt.Println("base64TxBytes:", encodedTxBytes)
			return nil
		},
	}
	cmd.Flags().Int(flagId, 0, "ID of the payment channel.")
	cmd.Flags().String(flagTo, "", "Recipient address of the payment channel")
	cmd.Flags().String(flagAmount, "", "Amount of coins to fund the paymetn channel with")
	return cmd
}

func ClosePaychanCmd(cdc *wire.Codec) *cobra.Command {
	flagState := "state"

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close a payment channel, given a state",
		Long:  "Close an existing payment channel with a state received from a sender. This signs it as the receiver before submitting to the blockchain.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// Get the sender-signed close tx
			state := viper.GetString(flagState)
			txBytes, err := base64.StdEncoding.DecodeString(state)
			if err != nil {
				return err
			}
			stdTx := auth.StdTx{}
			cdc.UnmarshalBinary(txBytes, &stdTx)

			// Sign close tx

			// ensure contxt has up to date account and sequence numbers
			ctx, err = Ensure(ctx)
			if err != nil {
				return err
			}
			// Sign message (asks user for password)
			_, sig, err := UserSignMsg(ctx, ctx.FromAddressName, stdTx.Msg)
			if err != nil {
				return err
			}

			// Append signature to close tx
			stdTx.Signatures = append(stdTx.Signatures, sig)
			// encode close tx
			txBytes, err = cdc.MarshalBinary(stdTx)
			if err != nil {
				return err
			}

			// Broadcast close tx to the blockchain

			res, err := ctx.BroadcastTx(txBytes)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}
	cmd.Flags().String(flagState, "", "State received from sender.")
	return cmd
}

// HELPER FUNCTIONS
// This is a partial refactor of cosmos-sdk/client/context.
// Existing API was awkard to use for paychans.

func Ensure(ctx context.CoreContext) (context.CoreContext, error) {

	ctx, err := context.EnsureAccountNumber(ctx)
	if err != nil {
		return ctx, err
	}
	// default to next sequence number if none provided
	ctx, err = context.EnsureSequence(ctx)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func UserSignMsg(ctx context.CoreContext, name string, msg sdk.Msg) (signMsg auth.StdSignMsg, stdSig auth.StdSignature, err error) {

	// TODO check how to handle non error return values on error. Returning empty versions doesn't seem right.

	passphrase, err := ctx.GetPassphraseFromStdin(name)
	if err != nil {
		return signMsg, stdSig, err
	}

	// build the Sign Messsage from the Standard Message
	chainID := ctx.ChainID
	if chainID == "" {
		return signMsg, stdSig, errors.Errorf("Chain ID required but not specified")
	}
	accnum := ctx.AccountNumber
	sequence := ctx.Sequence

	signMsg = auth.StdSignMsg{
		ChainID:        chainID,
		AccountNumbers: []int64{accnum},
		Sequences:      []int64{sequence},
		Msg:            msg,
		Fee:            auth.NewStdFee(ctx.Gas, sdk.Coin{}), // TODO run simulate to estimate gas?
	}

	keybase, err := keys.GetKeyBase()
	if err != nil {
		return signMsg, stdSig, err
	}

	// sign and build
	bz := signMsg.Bytes()

	sig, pubkey, err := keybase.Sign(name, passphrase, bz)
	if err != nil {
		return signMsg, stdSig, err
	}
	stdSig = auth.StdSignature{
		PubKey:        pubkey,
		Signature:     sig,
		AccountNumber: accnum,
		Sequence:      sequence,
	}

	return signMsg, stdSig, nil
}

func Build(cdc *wire.Codec, signMsg auth.StdSignMsg, sig auth.StdSignature) ([]byte, error) {
	tx := auth.NewStdTx(signMsg.Msg, signMsg.Fee, []auth.StdSignature{sig})
	return cdc.MarshalBinary(tx)
}

func EnsureSignBuild(ctx context.CoreContext, name string, msg sdk.Msg, cdc *wire.Codec) ([]byte, error) {
	//Ensure context has up to date account and sequence numbers
	ctx, err := Ensure(ctx)
	if err != nil {
		return nil, err
	}
	// Sign message (asks user for password)
	signMsg, sig, err := UserSignMsg(ctx, name, msg)
	if err != nil {
		return nil, err
	}
	// Create tx and marshal
	txBytes, err := Build(cdc, signMsg, sig)
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}
*/
