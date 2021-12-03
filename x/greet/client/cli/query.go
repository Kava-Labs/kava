package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/kava-labs/kava/x/greet/types"
	"github.com/spf13/cobra"
)

// this is the parent query command for the greet module everytime we add a new command we will register it here
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group todos queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdListGreetings())
	cmd.AddCommand(CmdShowGreeting())

	return cmd
}


// build the list greet command function 
func CmdListGreetings() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-greetings",
		Short: "list all greetings",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllGreetRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.GreetAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// build the show greet command function 
func CmdShowGreeting() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-greeting [id]",
		Short: "shows a greeting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetGreetRequest{
				Id: args[0],
			}

			res, err := queryClient.Greet(context.Background(), params)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res) 
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}