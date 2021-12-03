package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: types.ModuleName,
		Short: fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing: true,
		SuggestionsMinimumDistance: 2,
		RunE: client.ValidateCmd,
	}
	cmd.AddCommand(CmdCreateGreeting())
	return cmd
}


func CmdCreateGreeting() *cobra.Command {
	cmd:= &cobra.Command{
		Use: "create-greeting [message]",
		Short: "creates a new greetings",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := string(args[0])
		
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err 
			}
	
			msg := types.NewMsgCreateGreet(clientCtx.GetFromAddress().String(), string(message))
		
			if err := msg.ValidateBasic(); err != nil {
				return err 
			}
			
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}