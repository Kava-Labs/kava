package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"

	"github.com/kava-labs/kava/internal/x/paychan"
)

// list of functions that return pointers to cobra commands
// No local storage needed for cli acting as a sender

func CreateChannelCmd(cdc *wire.Codec) *cobra.Command {
	flagTo := "to"
	flagCoins := "amount"

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new payment channel",
		Long:  "Create a new unidirectional payment channel from a local address to a remote address, funded with some amount of coins. These coins are removed from the sender account and put into the channel.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a tx and cli "contexts": structs populated with info from common flags.
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// Get sender address
			sender, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			// Get receiver address
			toStr := viper.GetString(flagTo)
			receiver, err := sdk.AccAddressFromBech32(toStr)
			if err != nil {
				return err
			}

			// Get channel funding amount
			coinsString := viper.GetString(flagCoins)
			coins, err := sdk.ParseCoins(coinsString)
			if err != nil {
				return err
			}

			// Create the create channel msg to send
			msg := paychan.MsgCreate{
				Participants: [2]sdk.AccAddress{sender, receiver},
				Coins:        coins,
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// Build and sign the transaction, then broadcast to the blockchain
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagTo, "", "Recipient address of the payment channel.")
	cmd.Flags().String(flagCoins, "", "Amount of coins to fund the payment channel with.")
	return cmd
}

func GeneratePaymentCmd(cdc *wire.Codec) *cobra.Command {
	flagId := "chan-id"
	flagReceiverAmount := "rec-amt" // amount the receiver should received on closing the channel
	flagSenderAmount := "sen-amt"
	flagPaymentFile := "filename"

	cmd := &cobra.Command{
		Use:   "pay",
		Short: "Generate a new payment.", // TODO descriptions
		Long:  "Generate a payment file (json) to send to the receiver as a payment.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a cli "context": struct populated with info from common flags.
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// Get the paychan id
			id := paychan.ChannelID(viper.GetInt64(flagId)) // TODO make this default to pulling id from chain

			// Get channel receiver amount
			senderCoins, err := sdk.ParseCoins(viper.GetString(flagSenderAmount))
			if err != nil {
				return err
			}
			// Get channel receiver amount
			receiverCoins, err := sdk.ParseCoins(viper.GetString(flagReceiverAmount))
			if err != nil {
				return err
			}

			// create close paychan msg
			update := paychan.Update{
				ChannelID: id,
				Payout:    paychan.Payout{senderCoins, receiverCoins},
				// empty sigs
			}

			// Sign the update as the sender
			keybase, err := keys.GetKeyBase()
			if err != nil {
				return err
			}
			name := cliCtx.FromAddressName
			passphrase, err := keys.GetPassphrase(cliCtx.FromAddressName)
			if err != nil {
				return err
			}
			bz := update.GetSignBytes()

			sig, pubKey, err := keybase.Sign(name, passphrase, bz)
			if err != nil {
				return err
			}
			update.Sigs = [1]paychan.UpdateSignature{{
				PubKey:          pubKey,
				CryptoSignature: sig,
			}}

			// Write out the update
			jsonUpdate, err := wire.MarshalJSONIndent(cdc, update)
			if err != nil {
				return err
			}
			paymentFile := viper.GetString(flagPaymentFile)
			err = ioutil.WriteFile(paymentFile, jsonUpdate, 0644)
			if err != nil {
				return err
			}
			fmt.Printf("Written payment out to %v.\n", paymentFile)

			return nil
		},
	}
	cmd.Flags().Int(flagId, 0, "ID of the payment channel.")
	cmd.Flags().String(flagSenderAmount, "", "Total coins to payout to sender on channel close.")
	cmd.Flags().String(flagReceiverAmount, "", "Total coins to payout to sender on channel close.")
	cmd.Flags().String(flagPaymentFile, "payment.json", "File name to write the payment into.")
	return cmd
}

func VerifyPaymentCmd(cdc *wire.Codec, paychanStoreName string) *cobra.Command {
	flagPaymentFile := "payment"

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a payment file.",
		Long:  "Verify that a received payment can be used to close a channel.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a cli "context": struct populated with info from common flags.
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// read in update
			bz, err := ioutil.ReadFile(viper.GetString(flagPaymentFile))
			if err != nil {
				// TODO add nice message about how to feed in stdin
				return err
			}
			// decode json
			var update paychan.Update
			cdc.UnmarshalJSON(bz, &update)

			// get the channel from the node
			res, err := cliCtx.QueryStore(paychan.GetChannelKey(update.ChannelID), paychanStoreName)
			if len(res) == 0 || err != nil {
				return errors.Errorf("channel with ID '%d' does not exist", update.ChannelID)
			}
			var channel paychan.Channel
			cdc.MustUnmarshalBinary(res, &channel)

			//verify
			verificationError := paychan.VerifyUpdate(channel, update)

			// print result
			if verificationError == nil {
				fmt.Printf("Payment is valid for channel '%d'.\n", update.ChannelID)
			} else {
				fmt.Printf("Payment is NOT valid for channel '%d'.\n", update.ChannelID)
				fmt.Println(verificationError)
			}
			return nil
		},
	}
	cmd.Flags().String(flagPaymentFile, "payment.json", "File name to read the payment from.")

	return cmd
}

func SubmitPaymentCmd(cdc *wire.Codec) *cobra.Command {
	flagPaymentFile := "payment"

	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a payment to the blockchain to close the channel.",
		Long:  fmt.Sprintf("Submit a payment to the blockchain to either close a channel immediately (if you are the receiver) or after a dispute period of %d blocks (if you are the sender).", paychan.ChannelDisputeTime),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a tx and cli "contexts": structs populated with info from common flags.
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// Get sender address
			submitter, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			// read in update
			bz, err := ioutil.ReadFile(viper.GetString(flagPaymentFile))
			if err != nil {
				return err
			}
			// decode json
			var update paychan.Update
			cdc.UnmarshalJSON(bz, &update)

			// Create the create channel msg to send
			msg := paychan.MsgSubmitUpdate{
				Update:    update,
				Submitter: submitter,
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// Build and sign the transaction, then broadcast to the blockchain
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagPaymentFile, "payment.json", "File to read the payment from.")
	return cmd
}

func GetChannelCmd(cdc *wire.Codec, paychanStoreName string) *cobra.Command {
	flagId := "chan-id"
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get info on a channel.",
		Long:  "Get the details of a non closed channel plus any submitted update waiting to be executed.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create a cli "context": struct populated with info from common flags.
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// Get channel ID
			id := paychan.ChannelID(viper.GetInt64(flagId))

			// Get the channel from the node
			res, err := cliCtx.QueryStore(paychan.GetChannelKey(id), paychanStoreName)
			if len(res) == 0 || err != nil {
				return errors.Errorf("channel with ID '%d' does not exist", id)
			}
			var channel paychan.Channel
			cdc.MustUnmarshalBinary(res, &channel)

			// Convert the channel to a json object for pretty printing
			jsonChannel, err := wire.MarshalJSONIndent(cdc, channel)
			if err != nil {
				return err
			}
			// print out json channel
			fmt.Println(string(jsonChannel))

			// Get any submitted updates from the node
			res, err = cliCtx.QueryStore(paychan.GetSubmittedUpdateKey(id), paychanStoreName)
			if err != nil {
				return err
			}
			// Print out the submitted update if it exists
			if len(res) != 0 {
				var submittedUpdate paychan.SubmittedUpdate
				cdc.MustUnmarshalBinary(res, &submittedUpdate)

				// Convert the submitted update to a json object for pretty printing
				jsonSU, err := wire.MarshalJSONIndent(cdc, submittedUpdate)
				if err != nil {
					return err
				}
				// print out json submitted update
				fmt.Println(string(jsonSU))
			}
			return nil
		},
	}
	cmd.Flags().Int(flagId, 0, "ID of the payment channel.")
	return cmd
}
