package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction cli commands for the incentive module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	incentiveTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "transaction commands for the incentive module",
	}

	incentiveTxCmd.AddCommand(flags.PostCommands(
		getCmdClaim(cdc),
	)...)

	return incentiveTxCmd

}

func getCmdClaim(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim [owner] [denom]",
		Short: "claim rewards for owner and denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim any outstanding rewards owned by owner for the input denom,

			Example:
			$ %s tx %s claim kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw bnb
		`, version.ClientName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimReward(owner, args[1])
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
