package main

import (
	"errors"
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
	"github.com/kava-labs/kava/cmd/kvoracle/feed"
)

// FlagRPCURL specifies the url for kava's rpc
const FlagRPCURL = "rpc-url"

var appCodec *amino.Codec

func main() {
	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	config.Seal()

	appCodec = app.MakeCodec()

	DefaultCLIHome := os.ExpandEnv("$HOME/.kvcli")

	// Add (--chain-id, --rpc-url) to persistent flags and mark them required
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentFlags().String(FlagRPCURL, "", "RPC URL of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		startPriceFeedCmd(),
	)

	executor := cli.PrepareMainCmd(rootCmd, "KVORACLE", DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		log.Fatal("failed executing CLI command", err)
	}
}

var rootCmd = &cobra.Command{
	Use:          "kvoracle",
	Short:        "Automatic price feed",
	SilenceUsage: true,
}

func startPriceFeedCmd() *cobra.Command {
	startPriceFeedCmd := &cobra.Command{
		Use:     "start [oracle-moniker] [coin1, coin2] [interval-minutes] --rpc-url=[rpc-url] --chain-id=[chain-id]",
		Short:   "Starts an oracle that automatically updates kava's price feed",
		Args:    cobra.ExactArgs(3),
		Example: "kvoracle start vlad bitcoin,kava 30 --rpc-url=tcp://localhost:26657 --chain-id=testing",
		RunE:    RunStartPriceFeedCmd,
	}

	return startPriceFeedCmd
}

// RunStartPriceFeedCmd runs the RunStartPriceFeed cmd
func RunStartPriceFeedCmd(cmd *cobra.Command, args []string) error {
	// Parse RPC URL
	rpcURL := viper.GetString(FlagRPCURL)
	if strings.TrimSpace(rpcURL) == "" {
		return errors.New("Must specify an 'rpc-url'")
	}

	// Parse chain's ID
	chainID := viper.GetString(client.FlagChainID)
	if strings.TrimSpace(chainID) == "" {
		return errors.New("Must specify a 'chain-id'")
	}

	// Parse the oracle's moniker
	oracleFrom := args[0]

	// Parse our coins
	coins := strings.Split(args[1], ",")
	if 1 > len(coins) {
		return errors.New("Must specify at least one coin")
	}

	// Parse the interval in minutes
	interval, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}
	if interval < 30 {
		return errors.New("Must specify an interval of 30 seconds or longer")
	}

	// Get the oracle's name and account address using their moniker
	accAddress, oracleName, sdkErr := context.GetFromFields(oracleFrom, false)
	if sdkErr != nil {
		return sdkErr
	}

	// Get the oracle's passphrase using their moniker
	passphrase, sdkErr := keys.GetPassphrase(oracleFrom)
	if sdkErr != nil {
		return sdkErr
	}

	// Test passphrase is correct
	_, sdkErr = authtypes.MakeSignature(nil, oracleFrom, passphrase, authtypes.StdSignMsg{})
	if sdkErr != nil {
		return sdkErr
	}

	// Set up our CLIContext
	cliCtx := context.NewCLIContext().
		WithCodec(appCodec).
		WithFromAddress(accAddress).
		WithFromName(oracleName)

	// Schedule cron for price collection and posting
	gocron.Every(uint64(interval)).Seconds().Do(feed.ExecutePostingIteration, coins, accAddress, chainID, appCodec, oracleName, passphrase, cliCtx, rpcURL)
	<-gocron.Start()
	gocron.Clear()

	return nil
}

func initConfig(cmd *cobra.Command) error {
	err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID))
	if err != nil {
		return err
	}
	return viper.BindPFlag(FlagRPCURL, cmd.PersistentFlags().Lookup(FlagRPCURL))
}
