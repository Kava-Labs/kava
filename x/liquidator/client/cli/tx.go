package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/liquidator/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "liquidator",
		Short: "liquidator transactions subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		GetCmdSeizeAndStartCollateralAuction(cdc),
		GetCmdStartDebtAuction(cdc),
	)...)

	return txCmd
}

// GetCmdSeizeAndStartCollateralAuction seize funds from a CDP and send to auction
func GetCmdSeizeAndStartCollateralAuction(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seize [cdp-owner] [collateral-denom]",
		Short: "",
		Long: `Seize a fixed amount of collateral and debt from a CDP then start an auction with the collateral.
The amount of collateral seized is given by the 'AuctionSize' module parameter or, if there isn't enough collateral in the CDP, all the CDP's collateral is seized.
Debt is seized in proportion to the collateral seized so that the CDP stays at the same collateral to debt ratio.
A 'forward-reverse' auction is started selling the seized collateral for some stable coin, with a maximum bid of stable coin set to equal the debt seized.
As this is a forward-reverse auction type, if the max stable coin is bid then bidding continues by bidding down the amount of collateral taken by the bidder. At the end, extra collateral is returned to the original CDP owner.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			// Validate inputs
			sender := cliCtx.GetFromAddress()
			cdpOwner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			denom := args[1]
			// TODO validate denom?

			// Prepare and send message
			msg := types.MsgSeizeAndStartCollateralAuction{
				Sender:          sender,
				CdpOwner:        cdpOwner,
				CollateralDenom: denom,
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

func GetCmdStartDebtAuction(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "start a debt auction, minting gov coin to cover debt",
		Long:  "Start a reverse auction, selling off minted gov coin to raise a fixed amount of stable coin.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()

			// Prepare and send message
			msg := types.MsgStartDebtAuction{
				Sender: sender,
			}

			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			// TODO print out results like auction ID?
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
