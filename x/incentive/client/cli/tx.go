package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/incentive/types"
)

const (
	multiplierFlag      = "multiplier"
	multiplierFlagShort = "m"
)

// GetTxCmd returns the transaction cli commands for the incentive module
func GetTxCmd() *cobra.Command {
	incentiveTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "transaction commands for the incentive module",
	}

	cmds := []*cobra.Command{
		getCmdClaimCdp(),
		getCmdClaimHard(),
		getCmdClaimDelegator(),
		getCmdClaimSwap(),
		getCmdClaimSavings(),
		getCmdClaimEarn(),
	}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	incentiveTxCmd.AddCommand(cmds...)

	return incentiveTxCmd
}

func getCmdClaimCdp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "claim-cdp [multiplier]",
		Short:   "claim USDX minting rewards using a given multiplier",
		Long:    `Claim sender's outstanding USDX minting rewards using a given multiplier.`,
		Example: fmt.Sprintf(`  $ %s tx %s claim-cdp large`, version.AppName, types.ModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]

			msg := types.NewMsgClaimUSDXMintingReward(sender.String(), multiplier)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}

	return cmd
}

func getCmdClaimHard() *cobra.Command {
	var denomsToClaim map[string]string

	cmd := &cobra.Command{
		Use:   "claim-hard",
		Short: "claim sender's Hard module rewards using given multipliers",
		Long:  `Claim sender's outstanding Hard rewards for deposit/borrow using given multipliers`,
		Example: strings.Join([]string{
			fmt.Sprintf(`  $ %s tx %s claim-hard --%s hard=large --%s ukava=small`, version.AppName, types.ModuleName, multiplierFlag, multiplierFlag),
			fmt.Sprintf(`  $ %s tx %s claim-hard --%s hard=large,ukava=small`, version.AppName, types.ModuleName, multiplierFlag),
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			selections := types.NewSelectionsFromMap(denomsToClaim)

			msg := types.NewMsgClaimHardReward(sender.String(), selections)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	cmd.Flags().StringToStringVarP(&denomsToClaim, multiplierFlag, multiplierFlagShort, nil, "specify the denoms to claim, each with a multiplier lockup")
	if err := cmd.MarkFlagRequired(multiplierFlag); err != nil {
		panic(err)
	}
	return cmd
}

func getCmdClaimDelegator() *cobra.Command {
	var denomsToClaim map[string]string

	cmd := &cobra.Command{
		Use:   "claim-delegator",
		Short: "claim sender's non-sdk delegator rewards using given multipliers",
		Long:  `Claim sender's outstanding delegator rewards using given multipliers`,
		Example: strings.Join([]string{
			fmt.Sprintf(`  $ %s tx %s claim-delegator --%s hard=large --%s swp=small`, version.AppName, types.ModuleName, multiplierFlag, multiplierFlag),
			fmt.Sprintf(`  $ %s tx %s claim-delegator --%s hard=large,swp=small`, version.AppName, types.ModuleName, multiplierFlag),
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			selections := types.NewSelectionsFromMap(denomsToClaim)

			msg := types.NewMsgClaimDelegatorReward(sender.String(), selections)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	cmd.Flags().StringToStringVarP(&denomsToClaim, multiplierFlag, multiplierFlagShort, nil, "specify the denoms to claim, each with a multiplier lockup")
	if err := cmd.MarkFlagRequired(multiplierFlag); err != nil {
		panic(err)
	}
	return cmd
}

func getCmdClaimSwap() *cobra.Command {
	var denomsToClaim map[string]string

	cmd := &cobra.Command{
		Use:   "claim-swap",
		Short: "claim sender's swap rewards using given multipliers",
		Long:  `Claim sender's outstanding swap rewards using given multipliers`,
		Example: strings.Join([]string{
			fmt.Sprintf(`  $ %s tx %s claim-swap --%s swp=large --%s ukava=small`, version.AppName, types.ModuleName, multiplierFlag, multiplierFlag),
			fmt.Sprintf(`  $ %s tx %s claim-swap --%s swp=large,ukava=small`, version.AppName, types.ModuleName, multiplierFlag),
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			selections := types.NewSelectionsFromMap(denomsToClaim)

			msg := types.NewMsgClaimSwapReward(sender.String(), selections)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	cmd.Flags().StringToStringVarP(&denomsToClaim, multiplierFlag, multiplierFlagShort, nil, "specify the denoms to claim, each with a multiplier lockup")
	if err := cmd.MarkFlagRequired(multiplierFlag); err != nil {
		panic(err)
	}
	return cmd
}

func getCmdClaimSavings() *cobra.Command {
	var denomsToClaim map[string]string

	cmd := &cobra.Command{
		Use:   "claim-savings",
		Short: "claim sender's savings rewards using given multipliers",
		Long:  `Claim sender's outstanding savings rewards using given multipliers`,
		Example: strings.Join([]string{
			fmt.Sprintf(`  $ %s tx %s claim-savings --%s swp=large --%s ukava=small`, version.AppName, types.ModuleName, multiplierFlag, multiplierFlag),
			fmt.Sprintf(`  $ %s tx %s claim-savings --%s swp=large,ukava=small`, version.AppName, types.ModuleName, multiplierFlag),
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			selections := types.NewSelectionsFromMap(denomsToClaim)

			msg := types.NewMsgClaimSavingsReward(sender.String(), selections)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	cmd.Flags().StringToStringVarP(&denomsToClaim, multiplierFlag, multiplierFlagShort, nil, "specify the denoms to claim, each with a multiplier lockup")
	if err := cmd.MarkFlagRequired(multiplierFlag); err != nil {
		panic(err)
	}
	return cmd
}

func getCmdClaimEarn() *cobra.Command {
	var denomsToClaim map[string]string

	cmd := &cobra.Command{
		Use:   "claim-earn",
		Short: "claim sender's earn rewards using given multipliers",
		Long:  `Claim sender's outstanding earn rewards using given multipliers`,
		Example: strings.Join([]string{
			fmt.Sprintf(`  $ %s tx %s claim-earn --%s swp=large --%s ukava=small`, version.AppName, types.ModuleName, multiplierFlag, multiplierFlag),
			fmt.Sprintf(`  $ %s tx %s claim-earn --%s swp=large,ukava=small`, version.AppName, types.ModuleName, multiplierFlag),
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			selections := types.NewSelectionsFromMap(denomsToClaim)

			msg := types.NewMsgClaimEarnReward(sender.String(), selections)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), &msg)
		},
	}
	cmd.Flags().StringToStringVarP(&denomsToClaim, multiplierFlag, multiplierFlagShort, nil, "specify the denoms to claim, each with a multiplier lockup")
	if err := cmd.MarkFlagRequired(multiplierFlag); err != nil {
		panic(err)
	}
	return cmd
}
