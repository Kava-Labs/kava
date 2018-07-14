package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/spf13/cobra"
	//"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/kava-labs/kava/internal/x/paychan"
)

// list of functions that return pointers to cobra commands
// No local storage needed for cli acting as a sender

// create paychan
// close paychan
// get paychan(s)
// send paychan payment
// get balance from receiver

// minimum
// create paychan (sender signs)
// create state update (sender signs) (just a half signed close tx, (json encoded?))
// close paychan (receiver signs) (provide state update as arg)

func CreatePaychanCmd(cdc *wire.Codec) *cobra.Command {
	flagTo := "to"
	flagAmount := "amount"

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new payment channel",
		Long:  "Create a new payment channel from a local address to a remote address, funded with some amount of coins. These coins are removed from the sender account and put into the payment channel.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// get args: from, to, amount
			// create a "client context" stuct populated with info from common flags
			ctx := context.NewCoreContextFromViper()
			// ChainID:         chainID,
			// Height:          viper.GetInt64(client.FlagHeight),
			// Gas:             viper.GetInt64(client.FlagGas),
			// TrustNode:       viper.GetBool(client.FlagTrustNode),
			// FromAddressName: viper.GetString(client.FlagName),
			// NodeURI:         nodeURI,
			// AccountNumber:   viper.GetInt64(client.FlagAccountNumber),
			// Sequence:        viper.GetInt64(client.FlagSequence),
			// Client:          rpc,
			// Decoder:         nil,
			// AccountStore:    "acc",

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
				sender:   senderAddress,
				receiver: receiverAddress,
				amount:   amount,
			}
			// Build and sign the transaction, then broadcast to Tendermint
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
		},
	}
	cmd.Flags().String(flagTo, "", "Recipient address of the payment channel")
	cmd.Flags().String(flagAmount, "", "Amount of coins to fund the paymetn channel with")
	return cmd
}

func CreateNewStateCmd(cdc *wire.Codec) *cobra.Command {
	flagId := "id"
	flagTo := "to"
	flagAmount := "amount"

	cmd := &cobra.Command{
		Use:   "localstate",
		Short: "Create a payment channel claim",
		Long:  "",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// sender(name) receiver id

			// create a "client context" stuct populated with info from common flags
			ctx := context.NewCoreContextFromViper()
			// Get sender adress
			senderAddress, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}
			// Get id
			id := viper.GetInt(flagId)
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

			// create MsgClose

			msg := paychan.MsgClose{
				sender:         senderAddress,
				receiver:       receiverAddress,
				id:             id,
				receiverAmount: amount,
			}

			// half sign it

			txBytes, err := EnsureSignBuild(ctx, ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			// print it out

			fmt.Println(txBytes)

		},
	}
	cmd.Flags().Int(flagId, 0, "ID of the payment channel.")
	cmd.Flags().String(flagTo, "", "Recipient address of the payment channel")
	cmd.Flags().String(flagAmount, "", "Amount of coins to fund the paymetn channel with")
	return cmd
}

// sign and build the transaction from the msg
func EnsureSignBuild(ctx context.CoreContext, name string, msg sdk.Msg, cdc *wire.Codec) ([]byte, error) {

	ctx, err = EnsureAccountNumber(ctx)
	if err != nil {
		return nil, err
	}
	// default to next sequence number if none provided
	ctx, err = EnsureSequence(ctx)
	if err != nil {
		return nil, err
	}

	passphrase, err := ctx.GetPassphraseFromStdin(name)
	if err != nil {
		return nil, err
	}

	txBytes, err := ctx.SignAndBuild(name, passphrase, msg, cdc)
	if err != nil {
		return nil, err
	}

	return txBytes
}

func ClosePaychanCmd(cdc *wire.Codec) *cobra.Command {
	flagId := "id"
	flagTo := "to"
	flagState := "state"

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Close a payment channel",
		Long:  "",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// get sender, reciver, id
			// get state
			// sign the state tx with receiver
			// broadcast to tendermint
		},
	}
	//cmd.Flags().String(flagTo, "", "Recipient address of the payment channel")
	//cmd.Flags().String(flagAmount, "", "Amount of coins to fund the paymetn channel with")
	return cmd
}
