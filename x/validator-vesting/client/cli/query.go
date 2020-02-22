package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/x/cdp/types"
)


// GetQueryValidatorVesting returns the cli query commands for this module
func GetQueryValidatorVesting(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group nameservice queries under a subcommand
	cdpQueryCmd := &cobra.Command{
		Use:   "cdp",
		Short: "Querying commands for the cdp module",
	}
