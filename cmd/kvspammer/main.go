package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jasonlvhit/gocron"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/cmd/kvspammer/spammer"
)

// FlagRPCURL specifies the url for kava's rpc
const FlagRPCURL = "rpc-url"

// FlagFrom specifies a moniker of an address on kava
const FlagFrom = "from"

var appCodec *amino.Codec

func main() {
	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	config.Seal()

	appCodec = app.MakeCodec()

	DefaultCLIHome := os.ExpandEnv("$HOME/.kvcli")

	// Add [--from], [--chain-id], [--rpc-url] to persistent flags and mark them required
	rootCmd.PersistentFlags().String(FlagFrom, "", "Moniker of address on Kava blockchain")
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentFlags().String(FlagRPCURL, "", "RPC URL of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct cdp command
	cdpCmd.AddCommand(
		generateCDPsCmd(),
	)

	// Construct root command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		cdpCmd,
	)

	executor := cli.PrepareMainCmd(rootCmd, "KVSPAMMER", DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		log.Fatal("failed executing CLI command", err)
	}
}

var rootCmd = &cobra.Command{
	Use:          "kvspammer",
	Short:        "Spam bot that creates transactions on the Kava blockchain",
	SilenceUsage: true,
}

var cdpCmd = &cobra.Command{
	Use:   "cdp",
	Short: "CDP subcommands",
}

func generateCDPsCmd() *cobra.Command {
	generateCDPsCmd := &cobra.Command{
		Use:     "gencdps [collateral-denom] [debt-denom] [max-collateral] [interval-seconds] --from=[moniker] --rpc-url=[rpc-url] --chain-id=[chain-id]",
		Short:   "Initalizes a feed which generates random CDPs within the parameterized bounds",
		Args:    cobra.ExactArgs(4),
		Example: "kvspammer cdp gencdps btc usdx 50 20 --from=vlad --rpc-url=tcp://localhost:26657 --chain-id=testing",
		RunE:    RunGenerateCDPsCmd,
	}

	return generateCDPsCmd
}

// RunGenerateCDPsCmd runs the generate CDPs command
func RunGenerateCDPsCmd(cmd *cobra.Command, args []string) error {
	// Parse from moniker URL
	from := viper.GetString(FlagFrom)
	if strings.TrimSpace(from) == "" {
		return errors.New("Must specify a 'from' moniker")
	}

	// Parse chain's ID
	chainID := viper.GetString(client.FlagChainID)
	if strings.TrimSpace(chainID) == "" {
		return errors.New("Must specify a 'chain-id'")
	}

	// Parse RPC URL
	rpcURL := viper.GetString(FlagRPCURL)
	if strings.TrimSpace(rpcURL) == "" {
		return errors.New("Must specify a 'rpc-url'")
	}

	// TODO: Validate that this denom exists in the app
	collateralDenom := args[0]
	if len(collateralDenom) == 0 {
		return errors.New("Must specify a valid collateral denom")
	}

	principalDenom := args[1]
	if len(principalDenom) == 0 {
		return errors.New("Must specify a valid debt denom")
	}

	maxCollateral, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		fmt.Printf("Invalid max collateral: %s \n", string(args[2]))
		return err
	}

	interval, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return err
	}
	if interval < 10 {
		fmt.Printf("Invalid internal %s (minimum 10 seconds)\n", string(interval))
	}

	// Get the spammer's name and account address using their moniker
	accAddress, _, sdkErr := context.GetFromFields(from, false)
	if sdkErr != nil {
		return sdkErr
	}

	// Get the spammer's passphrase using their moniker
	passphrase, sdkErr := keys.GetPassphrase(from)
	if sdkErr != nil {
		return sdkErr
	}

	// Test passphrase is correct
	_, sdkErr = authtypes.MakeSignature(nil, from, passphrase, authtypes.StdSignMsg{})
	if sdkErr != nil {
		return sdkErr
	}

	// Set up our CLIContext
	cliCtx := context.NewCLIContext().
		WithCodec(appCodec).
		WithFromAddress(accAddress).
		WithFromName(from)

	_ = cliCtx
	_ = maxCollateral

	gocron.Every(uint64(interval)).Seconds().Do(
		spammer.SpamTxCDP(
			rpcURL,
			chainID,
			from,
			passphrase,
			collateralDenom,
			principalDenom,
			maxCollateral,
			appCodec,
			&cliCtx,
			accAddress,
		),
	)

	// gocron.Every(uint64(interval)).Seconds().Do(func() { fmt.Println("execute") })

	<-gocron.Start()
	gocron.Clear()

	return nil
}

func initConfig(cmd *cobra.Command) error {
	err := viper.BindPFlag(FlagFrom, cmd.PersistentFlags().Lookup(FlagFrom))
	if err != nil {
		return err
	}
	err = viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID))
	if err != nil {
		return err
	}
	return viper.BindPFlag(FlagRPCURL, cmd.PersistentFlags().Lookup(FlagRPCURL))
}
