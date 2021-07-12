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
	)...)

	return incentiveTxCmd
}

func getCmdClaimCdp(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim-cdp [multiplier]",
		Short: "claim CDP rewards using a given multiplier",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim sender's outstanding CDP rewards using a given multiplier

			Example:
			$ %s tx %s claim-cdp large
		`, version.ClientName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]

			msg := types.NewMsgClaimUSDXMintingReward(sender, multiplier)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdClaimCdpVVesting(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim-cdp-vesting [multiplier] [receiver]",
		Short: "claim CDP rewards using a given multiplier on behalf of a validator vesting account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim sender's outstanding CDP rewards on behalf of a validator vesting using a given multiplier

			Example:
			$ %s tx %s claim-cdp-vesting large kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
		`, version.ClientName, types.ModuleName),
		),
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

			msg := types.NewMsgClaimUSDXMintingRewardVVesting(sender, receiver, multiplier)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdClaimHardVVesting(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim-hard-vesting [multiplier] [receiver]",
		Short: "claim Hard module rewards on behalf of a validator vesting account using a given multiplier",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim sender's outstanding Hard rewards on behalf of a validator vesting account for deposit/borrow/delegate using given multiplier

			Example:
			$ %s tx %s claim-hard-vesting large kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
		`, version.ClientName, types.ModuleName),
		),
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

			msg := types.NewMsgClaimHardRewardVVesting(sender, receiver, multiplier)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdClaimHard(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim-hard [multiplier]",
		Short: "claim sender's Hard module rewards using a given multiplier",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim sender's outstanding Hard rewards for deposit/borrow/delegate using given multiplier

			Example:
			$ %s tx %s claim-hard large
		`, version.ClientName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]

			msg := types.NewMsgClaimHardReward(sender, multiplier)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdClaimDelegator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim-delegator [multiplier] [denoms to claim]",
		Short: "claim sender's delegator rewards using a given multiplier",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim sender's outstanding delegator rewards using given multiplier.
			Optionally claim only certain denoms from the rewards. Specifying none will claim all of them.

			Example:
			$ %s tx %s claim-delegator large swap,hard
		`, version.ClientName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			sender := cliCtx.GetFromAddress()
			multiplier := args[0]
			denomsToClaim := strings.Split(args[1], ",")

			msg := types.NewMsgClaimDelegatorReward(sender, multiplier, denomsToClaim)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func getCmdClaimDelegatorVVesting(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim-delegator-vesting [multiplier] [receiver]",
		Short: "claim delegator rewards on behalf of a validator vesting account using a given multiplier",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim sender's outstanding delegator rewards on behalf of a validator vesting account using given multiplier

			Example:
			$ %s tx %s claim-delegator-vesting large kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
		`, version.ClientName, types.ModuleName),
		),
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

			msg := types.NewMsgClaimDelegatorRewardVVesting(sender, receiver, multiplier)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
