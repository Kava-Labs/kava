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
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		getCmdClaimHard(cdc),
	)...)

	return incentiveTxCmd

}

func getCmdClaimCdp(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "claim-cdp [owner] [multiplier]",
		Short: "claim CDP rewards for cdp owner and collateral-type",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim any outstanding CDP rewards owned by owner for the input collateral-type and multiplier,

			Example:
			$ %s tx %s claim-cdp kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw large
		`, version.ClientName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			if !sender.Equals(owner) {
				return sdkerrors.Wrapf(types.ErrInvalidClaimOwner, "tx sender %s does not match claim owner %s", sender, owner)
			}

			msg := types.NewMsgClaimUSDXMintingReward(owner, args[1])
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
		Use:   "claim-hard [owner] [multiplier]",
		Short: "claim Hard rewards for deposit/borrow and delegating",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim owner's outstanding Hard rewards using given multiplier multiplier,

			Example:
			$ %s tx %s claim-hard kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw large
		`, version.ClientName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			sender := cliCtx.GetFromAddress()
			if !sender.Equals(owner) {
				return sdkerrors.Wrapf(types.ErrInvalidClaimOwner, "tx sender %s does not match claim owner %s", sender, owner)
			}

			msg := types.NewMsgClaimHardLiquidityProviderReward(owner, args[1])
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
