package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/community/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	communityTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "community module transactions subcommands",
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		getCmdFundCommunityPool(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	communityTxCmd.AddCommand(cmds...)

	return communityTxCmd
}

func getCmdFundCommunityPool() *cobra.Command {
	return &cobra.Command{
		Use:   "fund-community-pool [coins]",
		Short: "funds the community pool",
		Long:  "Fund community pool removes the listed coins from the sender's account and send them to the community module account.",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(
			`%s tx %s fund-community-module 10000000ukava --from <key>`, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgFundCommunityPool(clientCtx.GetFromAddress(), coins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
}
