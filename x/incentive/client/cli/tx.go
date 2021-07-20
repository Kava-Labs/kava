package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/incentive/types"
)

const multiplierFlag = "multiplier"
const multiplierFlagShort = "m"

// GetTxCmd returns the transaction cli commands for the incentive module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	incentiveTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "transaction commands for the incentive module",
	}

	incentiveTxCmd.AddCommand(flags.PostCommands(
		getCmdClaimCdp(cdc),
		getCmdClaimCdpVVesting(cdc),
		getCmdClaimHard(cdc),
		getCmdClaimHardVVesting(cdc),
		getCmdClaimDelegator(cdc),
		getCmdClaimDelegatorVVesting(cdc),
		getCmdClaimSwap(cdc),
		getCmdClaimSwapVVesting(cdc),
	)...)

	return incentiveTxCmd
}

func getCmdClaimCdp(cdc *codec.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "claim-cdp [multiplier]",
		Short:   "claim USDX minting rewards using a given multiplier",
		Long:    `Claim sender's outstanding USDX minting rewards using a given multiplier.`,
		Example: fmt.Sprintf(`  $ %s tx %s claim-cdp large`, version.ClientName, types.ModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]

			msg := types.NewMsgClaimUSDXMintingReward(sender, multiplier)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func getCmdClaimCdpVVesting(cdc *codec.Codec) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "claim-cdp-vesting [multiplier] [receiver]",
		Short: "claim USDX minting rewards using a given multiplier on behalf of a validator vesting account",
		Long: `Claim sender's outstanding USDX minting rewards on behalf of a validator vesting using a given multiplier.
A receiver address for the rewards is needed as validator vesting accounts cannot receive locked tokens.`,
		Example: fmt.Sprintf(`  $ %s tx %s claim-cdp-vesting large kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw`, version.ClientName, types.ModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]
			receiverStr := args[1]
			receiver, err := sdk.AccAddressFromBech32(receiverStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimUSDXMintingRewardVVesting(sender, receiver, multiplier)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

func getCmdClaimHard(cdc *codec.Codec) *cobra.Command {
	var denomsToClaim map[string]string

	cmd := &cobra.Command{
		Use:   "claim-hard",
		Short: "claim sender's Hard module rewards using given multipliers",
		Long:  `Claim sender's outstanding Hard rewards for deposit/borrow using given multipliers`,
		Example: strings.Join([]string{
			fmt.Sprintf(`  $ %s tx %s claim-hard --%s hard=large --%s ukava=small`, version.ClientName, types.ModuleName, multiplierFlag, multiplierFlag),
			fmt.Sprintf(`  $ %s tx %s claim-hard --%s hard=large,ukava=small`, version.ClientName, types.ModuleName, multiplierFlag),
		}, "\n"),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			selections := types.NewSelectionsFromMap(denomsToClaim)

			msg := types.NewMsgClaimHardReward(sender, selections...)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringToStringVarP(&denomsToClaim, multiplierFlag, multiplierFlagShort, nil, "specify the denoms to claim, each with a multiplier lockup")
	cmd.MarkFlagRequired(multiplierFlag)
	return cmd
}

// TODO improve help text for --multipliers flag

func getCmdClaimHardVVesting(cdc *codec.Codec) *cobra.Command {
	var denomsToClaim map[string]string

	cmd := &cobra.Command{
		Use:   "claim-hard-vesting [receiver]",
		Short: "claim Hard module rewards on behalf of a validator vesting account using given multipliers",
		Long: `Claim sender's outstanding hard supply/borrow rewards on behalf of a validator vesting account using given multipliers
A receiver address for the rewards is needed as validator vesting accounts cannot receive locked tokens.`,
		Example: strings.Join([]string{
			fmt.Sprintf("  $ %s tx %s claim-hard-vesting kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw --%s hard=large --%s ukava=small", version.ClientName, types.ModuleName, multiplierFlag, multiplierFlag),
			fmt.Sprintf("  $ %s tx %s claim-hard-vesting kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw --%s hard=large,ukava=small", version.ClientName, types.ModuleName, multiplierFlag),
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			receiver, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			selections := types.NewSelectionsFromMap(denomsToClaim)

			msg := types.NewMsgClaimHardRewardVVesting(sender, receiver, selections...)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringToStringVarP(&denomsToClaim, multiplierFlag, multiplierFlagShort, nil, "specify the denoms to claim, each with a multiplier lockup")
	cmd.MarkFlagRequired(multiplierFlag)

	return cmd
}

func getCmdClaimDelegator(cdc *codec.Codec) *cobra.Command {
	var denomsToClaim []string

	cmd := &cobra.Command{
		Use:   "claim-delegator [multiplier]",
		Short: "claim sender's delegator rewards using a given multiplier",
		Long: `Claim sender's outstanding delegator rewards using given multiplier.
Optionally claim only certain denoms from the rewards. Specifying none will claim all of them.`,
		Example: strings.Join([]string{
			fmt.Sprintf("  $ %s tx %s claim-delegator large", version.ClientName, types.ModuleName),
			fmt.Sprintf("  $ %s tx %s claim-delegator large --claim-only swp,hard", version.ClientName, types.ModuleName),
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]

			msg := types.NewMsgClaimDelegatorReward(sender, multiplier, denomsToClaim)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringSliceVar(&denomsToClaim, "claim-only", nil, "claim only these denoms, otherwise claim all denoms")

	return cmd
}

func getCmdClaimDelegatorVVesting(cdc *codec.Codec) *cobra.Command {
	var denomsToClaim []string

	cmd := &cobra.Command{
		Use:   "claim-delegator-vesting [multiplier] [receiver]",
		Short: "claim delegator rewards on behalf of a validator vesting account using a given multiplier",
		Long: `Claim sender's outstanding delegator rewards on behalf of a validator vesting account using given multiplier
A receiver address for the rewards is needed as validator vesting accounts cannot receive locked tokens.
Optionally claim only certain denoms from the rewards. Specifying none will claim all of them.`,
		Example: strings.Join([]string{
			fmt.Sprintf("  $ %s tx %s claim-delegator-vesting large kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw", version.ClientName, types.ModuleName),
			fmt.Sprintf("  $ %s tx %s claim-delegator-vesting small kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw --claim-only swp,hard", version.ClientName, types.ModuleName),
		}, "\n"),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]
			receiverStr := args[1]
			receiver, err := sdk.AccAddressFromBech32(receiverStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimDelegatorRewardVVesting(sender, receiver, multiplier, denomsToClaim)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringSliceVar(&denomsToClaim, "claim-only", nil, "claim only these denoms, otherwise claim all denoms")

	return cmd
}

func getCmdClaimSwap(cdc *codec.Codec) *cobra.Command {
	var denomsToClaim []string

	cmd := &cobra.Command{
		Use:   "claim-swap [multiplier]",
		Short: "claim sender's swap rewards using a given multiplier",
		Long: `Claim sender's outstanding swap rewards using given multiplier.
Optionally claim only certain denoms from the rewards. Specifying none will claim all of them.`,
		Example: strings.Join([]string{
			fmt.Sprintf("  $ %s tx %s claim-swap large", version.ClientName, types.ModuleName),
			fmt.Sprintf("  $ %s tx %s claim-swap large --claim-only swp,hard", version.ClientName, types.ModuleName),
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]

			msg := types.NewMsgClaimSwapReward(sender, multiplier, denomsToClaim)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringSliceVar(&denomsToClaim, "claim-only", nil, "claim only these denoms, otherwise claim all denoms")

	return cmd
}

func getCmdClaimSwapVVesting(cdc *codec.Codec) *cobra.Command {
	var denomsToClaim []string

	cmd := &cobra.Command{
		Use:   "claim-swap-vesting [multiplier] [receiver]",
		Short: "claim swap rewards on behalf of a validator vesting account using a given multiplier",
		Long: `Claim sender's outstanding swap rewards on behalf of a validator vesting account using given multiplier
A receiver address for the rewards is needed as validator vesting accounts cannot receive locked tokens.
Optionally claim only certain denoms from the rewards. Specifying none will claim all of them.`,
		Example: strings.Join([]string{
			fmt.Sprintf("  $ %s tx %s claim-swap-vesting large kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw", version.ClientName, types.ModuleName),
			fmt.Sprintf("  $ %s tx %s claim-swap-vesting small kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw --claim-only swp,hard", version.ClientName, types.ModuleName),
		}, "\n"),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]
			receiverStr := args[1]
			receiver, err := sdk.AccAddressFromBech32(receiverStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimSwapRewardVVesting(sender, receiver, multiplier, denomsToClaim)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringSliceVar(&denomsToClaim, "claim-only", nil, "claim only these denoms, otherwise claim all denoms")

	return cmd
}
