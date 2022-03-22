package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/kava-labs/kava/x/savings/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	savingsTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "savings transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{}

	for _, cmd := range cmds {
		flags.AddTxFlagsToCmd(cmd)
	}

	savingsTxCmd.AddCommand(cmds...)

	return savingsTxCmd
}
