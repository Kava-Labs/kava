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
	"github.com/kava-labs/kava/cmd/kvoracle/feed"
	"github.com/kava-labs/kava/cmd/kvoracle/txs"
)

var appCodec *amino.Codec

const (
	rpcURL  = "tcp://localhost:26657"
	chainID = "testing"
)

// FlagRPCURL specifies the url for kava's rpc
const FlagRPCURL = "rpc-url"

func init() {
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
		client.LineBreak,
		postAssetPriceCmd(),
		client.LineBreak,
		getInitPriceFeedCmd(),
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

func postAssetPriceCmd() *cobra.Command {
	postAssetPriceCmd := &cobra.Command{
		Use:     "postprice [oracle-moniker] [market] [price] --rpc-url=[rpc-url] --chain-id=[chain-id]",
		Short:   "Post the price of the base asset in a market",
		Args:    cobra.ExactArgs(3),
		Example: "kvoracle postprice testuser btc:usd 8105.93  --rpc-url=tcp://localhost:26657 --chain-id=testing",
		RunE:    RunPostAssetPriceCmd,
	}

	return postAssetPriceCmd
}

func getInitPriceFeedCmd() *cobra.Command {
	getInitPriceFeedCmd := &cobra.Command{
		Use:     "init [oracle-moniker] [coin1, coin2] [interval-minutes] --rpc-url=[rpc-url] --chain-id=[chain-id]",
		Short:   "Initialize an oracle that automatically updates kava's price feed",
		Args:    cobra.ExactArgs(3),
		Example: "kvoracle init testuser bitcoin,kava,ripple,binancecoin 5 --rpc-url=tcp://localhost:26657 --chain-id=testing",
		RunE:    RunInitPriceFeedCmd,
	}

	return getInitPriceFeedCmd
}

// RunPostAssetPriceCmd executes the getAssetPrice with the provided parameters
func RunPostAssetPriceCmd(cmd *cobra.Command, args []string) error {
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
	validatorFrom := args[0]

	// Parse the market code
	marketCode := args[1]

	// Parse the price
	price, err := sdk.NewDecFromStr(args[2])
	if err != nil {
		return err
	}

	// TODO: 'sdkErr' due to: https://github.com/cosmos/scaffold/pull/37

	// Get the validator's name and account address using their moniker
	accAddress, validatorName, sdkErr := context.GetFromFields(validatorFrom, false)
	if sdkErr != nil {
		return sdkErr
	}

	// Get the validator's passphrase using their moniker
	passphrase, sdkErr := keys.GetPassphrase(validatorFrom)
	if sdkErr != nil {
		return sdkErr
	}

	// Test passphrase is correct
	_, sdkErr = authtypes.MakeSignature(nil, validatorName, passphrase, authtypes.StdSignMsg{})
	if sdkErr != nil {
		return sdkErr
	}

	// Set up our CLIContext
	cliCtx := context.NewCLIContext().
		WithCodec(appCodec).
		WithFromAddress(accAddress).
		WithFromName(validatorName)

	// Build the msg
	msgPostPrice, sdkErr := txs.ConstructMsgPostPrice(accAddress, price, marketCode)
	if sdkErr != nil {
		return sdkErr
	}

	// Send tx containing msg to kava
	fmt.Printf("Posting price '%f' for %s...\n", msgPostPrice.Price, msgPostPrice.AssetCode)
	txRes, sdkErr := txs.SendTxPostPrice(chainID, appCodec, accAddress, validatorName, passphrase, cliCtx, &msgPostPrice, rpcURL)
	if sdkErr != nil {
		return sdkErr
	}

	fmt.Println("Tx hash:", txRes.TxHash)

	return nil
}

// RunInitPriceFeedCmd runs the InitPriceFeed Cmd cmd
func RunInitPriceFeedCmd(cmd *cobra.Command, args []string) error {
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
	if interval <= 1 {
		return errors.New("Must specify an interval of 2 minutes or longer")
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
	gocron.Every(uint64(interval)).Minutes().Do(feed.GetPricesAndPost, coins, accAddress, chainID, appCodec, oracleName, passphrase, cliCtx, rpcURL)
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

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
