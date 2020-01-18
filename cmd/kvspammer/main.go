package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

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
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/cmd/kvspammer/txs"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
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
		Use:     "gencdps [collateral-denom] [debt-denom] [max-collateral] --from=[moniker] --rpc-url=[rpc-url] --chain-id=[chain-id]",
		Short:   "Initalizes a feed which generates random CDPs within the parameterized bounds",
		Args:    cobra.ExactArgs(3),
		Example: "kvspammer cdp gencdps btc usdx 50 --from=vlad --rpc-url=tcp://localhost:26657 --chain-id=testing",
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

	maxCollateral, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		fmt.Printf("Invalid max collateral: %s \n", string(args[2]))
		return err
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

	// Get existing cdp
	cdp, found, err := txs.QueryCDP(appCodec, &cliCtx, accAddress, collateralDenom)
	fmt.Println("existing cdp:", found)
	if err != nil {
		fmt.Println(err)
	}

	// Get random collateral amount between (min, max)
	randSource := rand.New(rand.NewSource(int64(time.Now().Unix())))

	var msg []sdk.Msg

	if len(cdp.String()) < 100 { // TODO: Check this length against nil CDP
		// Create collateral and principal coin
		collateralAmount := sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, int(maxCollateral))))
		collateral := sdk.NewCoin(collateralDenom, collateralAmount)
		principal := sdk.NewCoin(principalDenom, collateralAmount.Quo(sdk.NewInt(2)))
		fmt.Printf("Creating new CDP. Collateral: %s, Principal: %s...\n", collateral, principal)
		msg = []sdk.Msg{cdptypes.NewMsgCreateCDP(accAddress, sdk.NewCoins(collateral), sdk.NewCoins(principal))}
	} else {
		fmt.Println("Current CDP:", cdp)

		// Get current values
		currCollateral := sdk.NewDec(int64(0))
		currPrincipal := sdk.NewDec(int64(0))
		currFees := sdk.NewDec(int64(0))

		// Error checking in case any value is empty
		if len(cdp.Collateral) > 0 {
			currCollateral = sdk.NewDec(cdp.Collateral[0].Amount.Int64())
		}
		if len(cdp.Principal) > 0 {
			currPrincipal = sdk.NewDec(cdp.Principal[0].Amount.Int64())
		}
		if len(cdp.AccumulatedFees) > 0 {
			currFees = sdk.NewDec(cdp.AccumulatedFees[0].Amount.Int64())
		}

		// Edge case 0 principal, 0 fees results in divide by 0
		var collateralizationRatio sdk.Dec
		if currPrincipal.Add(currFees).IsPositive() {
			collateralizationRatio = currCollateral.Quo(currPrincipal.Add(currFees))
		} else {
			// TODO: check this case
			collateralizationRatio = sdk.NewDec(int64(1000))
		}

		fmt.Println("Collateralization ratio:", collateralizationRatio)

		// TODO: Parameterize 220%
		if collateralizationRatio.GTE(sdk.NewDec(int64(220)).Quo(sdk.NewDec(int64(100)))) { // CDP's collateralization ratio is high
			if randSource.Int63()%2 == 0 {
				// Withdraw 1-20%
				coin := sdk.NewCoin(collateralDenom, sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, 20))))
				fmt.Printf("Attempting to withdraw %s collateral...\n", coin)
				msg = []sdk.Msg{cdptypes.NewMsgWithdraw(accAddress, accAddress, sdk.NewCoins(coin))}
			} else {
				// Draw principal 1-20%
				coin := sdk.NewCoin(principalDenom, sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, 20))))
				fmt.Printf("Attempting to draw %s principal...\n", coin)
				msg = []sdk.Msg{cdptypes.NewMsgDrawDebt(accAddress, collateralDenom, sdk.NewCoins(coin))}
			}
		} else { // CDP's collateralization ratio is low
			if randSource.Int63()%2 == 0 {
				// Deposit collateral 1-20%
				coin := sdk.NewCoin(collateralDenom, sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, 20))))
				fmt.Printf("Attempting to deposit %s collateral...\n", coin)
				msg = []sdk.Msg{cdptypes.NewMsgDeposit(accAddress, accAddress, sdk.NewCoins(coin))}
			} else {
				// Repay principal 1-20%
				coin := sdk.NewCoin(principalDenom, sdk.NewInt(int64(simulation.RandIntBetween(randSource, 1, 20))))
				fmt.Printf("Attempting to repay %s principal...\n", coin)
				msg = []sdk.Msg{cdptypes.NewMsgRepayDebt(accAddress, principalDenom, sdk.NewCoins(coin))}
			}
		}
	}

	// Send tx containing the msg
	txRes, sdkErr := txs.SendTxRPC(chainID, appCodec, accAddress, from, passphrase, cliCtx, msg, rpcURL)
	if sdkErr != nil {
		return sdkErr
	}

	fmt.Println("Tx hash:", txRes.TxHash)
	fmt.Println("Tx logs:", txRes.Logs)

	// TODO: Schedule cron for price collection and posting
	// gocron.Every(uint64(interval)).Minutes().Do(.GetPricesAndPost, coins, accAddress, chainID, appCodec, oracleName, passphrase, cliCtx, rpcURL)
	gocron.Every(uint64(2)).Minutes().Do(fmt.Println("here"))
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
