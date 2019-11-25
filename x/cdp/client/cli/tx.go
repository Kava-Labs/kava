package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/kava-labs/kava/x/cdp/types"

)

// GetTxCmd returns the transaction commands for this module
// TODO: Tests, see: https://github.com/cosmos/cosmos-sdk/blob/18de630d0ae1887113e266982b51c2bf1f662edb/x/staking/client/cli/tx_test.go
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	cdpTxCmd := &cobra.Command{
		Use:   "cdp",
		Short: "cdp transactions subcommands",
	}

	cdpTxCmd.AddCommand(client.PostCommands(
		GetCmdModifyCdp(cdc),
	)...)

	return cdpTxCmd
}

// GetCmdModifyCdp cli command for creating and modifying cdps.
func GetCmdModifyCdp(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "modifycdp [ownerAddress] [collateralType] [collateralChange] [debtChange]",
		Short: "create or modify a cdp",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			collateralChange, ok := sdk.NewIntFromString(args[2])
			if !ok {
				fmt.Printf("invalid collateral amount - %s \n", string(args[2]))
				return nil
			}
			debtChange, ok := sdk.NewIntFromString(args[3])
			if !ok {
				fmt.Printf("invalid debt amount - %s \n", string(args[3]))
				return nil
			}
			msg := types.NewMsgCreateOrModifyCDP(cliCtx.GetFromAddress(), args[1], collateralChange, debtChange)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
