package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/cmd/kvoracle/txs"
	pftypes "github.com/kava-labs/kava/x/pricefeed/types"
	// "github.com/kava-labs/kava/cmd/kvoracle/feed"
)

var appCodec *amino.Codec

// TODO: const FlagRPCURL = "rpc-url"

func init() {

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	appCodec = app.MakeCodec()

	DefaultCLIHome := os.ExpandEnv("$HOME/.kvcli")

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	// rootCmd.PersistentFlags().String(FlagRPCURL, "", "RPC URL of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		postAssetPriceCmd(),
		client.LineBreak,
		getAssetPriceCmd(),
		client.LineBreak,
		getInitPriceCollectionCmd(),
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
		Use:     "price post [asset] [account-address]",
		Short:   "",
		Args:    cobra.ExactArgs(2),
		Example: "kvfeed price btc:usd kava1302ukkphmpkqjm49prh7tf8y6vvslmrsp64f2m --chain-id=testing", //--chain-id=testing
		RunE:    RunPostAssetPriceCmd,
	}

	return postAssetPriceCmd
}

func getAssetPriceCmd() *cobra.Command {
	getAssetPriceCmd := &cobra.Command{
		Use:     "price get [asset]",
		Short:   "",
		Args:    cobra.ExactArgs(2),
		Example: "kvfeed price btc:usd kava1302ukkphmpkqjm49prh7tf8y6vvslmrsp64f2m --chain-id=testing", //--chain-id=testing
		RunE:    RunGetAssetPriceCmd,
	}

	return getAssetPriceCmd
}

func getInitPriceCollectionCmd() *cobra.Command {
	getInitPriceCollectionCmd := &cobra.Command{
		Use:     "init",
		Short:   "",
		Args:    cobra.ExactArgs(2),
		Example: "kvfeed price btc:usd kava1302ukkphmpkqjm49prh7tf8y6vvslmrsp64f2m --chain-id=testing", //--chain-id=testing
		RunE:    RunInitPriceCollectionCmd,
	}

	return getInitPriceCollectionCmd
}

// RunPostAssetPriceCmd executes the getAssetPrice with the provided parameters
func RunPostAssetPriceCmd(cmd *cobra.Command, args []string) error {
	// Parse chain's ID
	chainID := viper.GetString(client.FlagChainID)
	if strings.TrimSpace(chainID) == "" {
		return errors.New("Must specify a 'chain-id'")
	}

	// Parse the validator's moniker
	validatorFrom := args[1]

	// Parse Tendermint RPC URL
	rpcURL := viper.GetString(FlagRPCURL)

	if rpcURL != "" {
		_, err := url.Parse(rpcURL)
		if rpcURL != "" && err != nil {
			return fmt.Errorf("invalid RPC URL: %v", rpcURL)
		}
	}

	// Get the validator's name and account address using their moniker
	validatorAccAddress, validatorName, err := sdkContext.GetFromFields(validatorFrom, false)
	if err != nil {
		return err
	}
	// Convert the validator's account address into type ValAddress
	validatorAddress := sdk.ValAddress(validatorAccAddress)

	// Get the validator's passphrase using their moniker
	passphrase, err := keys.GetPassphrase(validatorFrom)
	if err != nil {
		return err
	}

	// Test passphrase is correct
	_, err = authtxb.MakeSignature(nil, validatorName, passphrase, authtxb.StdSignMsg{})
	if err != nil {
		return err
	}

	// Set up our CLIContext
	cliCtx := sdkContext.NewCLIContext().
		WithCodec(appCodec).
		WithFromAddress(sdk.AccAddress(validatorAddress)).
		WithFromName(validatorName)

	// Construct message
	addr, err := sdk.AccAddressFromBech32("kava1302ukkphmpkqjm49prh7tf8y6vvslmrsp64f2m")
	if err != nil {
		return err
	}
	expiry := time.Now().Add(1 * time.Hour)
	price, err := sdk.NewDecFromStr("8001.00")


	msgPostPrice := pftypes.NewMsgPostPrice(addr, "xrp", price, expiry)

	// txs.SendTxPostPrice()

	return nil

}

func RunGetAssetPriceCmd(cmd *cobra.Command, args []string) error {
	assetsRaw := args[0]
	if len(assetsRaw) < 1 {
		return errors.New("Must specify assets")
	}
	// GetAssetPrice
	return nil
}

func RunInitPriceCollectionCmd(cmd *cobra.Command, args []string) error {
	// TODO: Parse symbols []string

	// Get time, asset prices
	// now := time.Now().Format("15:04:05")
	// TODO: Import feed
	// assets := feed.GeckoPrices(symbols, "USD")

	// fmt.Println()
	// fmt.Println("Time: ", now)
	// fmt.Println("-------------")

	// // Print our coins
	// for _, asset := range assets {
	// 	fmt.Printf("%s: $%f\n", asset.Symbol, math.Round(asset.Price*1000)/1000)
	// }
	// fmt.Println()

	return nil
}

func initConfig(cmd *cobra.Command) error {
	return viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID))
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
