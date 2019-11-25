package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	cdpcmd "github.com/kava-labs/kava/x/cdp/client/cli"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

// NewModuleClient creates client for the module
func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	// Group nameservice queries under a subcommand
	cdpQueryCmd := &cobra.Command{
		Use:   "cdp",
		Short: "Querying commands for the cdp module",
	}

	cdpQueryCmd.AddCommand(client.GetCommands(
		cdpcmd.GetCmdGetCdp(mc.storeKey, mc.cdc),
		cdpcmd.GetCmdGetCdps(mc.storeKey, mc.cdc),
		cdpcmd.GetCmdGetUnderCollateralizedCdps(mc.storeKey, mc.cdc),
		cdpcmd.GetCmdGetParams(mc.storeKey, mc.cdc),
	)...)

	return cdpQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	cdpTxCmd := &cobra.Command{
		Use:   "cdp",
		Short: "cdp transactions subcommands",
	}

	cdpTxCmd.AddCommand(client.PostCommands(
		cdpcmd.GetCmdModifyCdp(mc.cdc),
	)...)

	return cdpTxCmd
}
