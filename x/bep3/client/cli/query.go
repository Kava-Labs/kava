package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/bep3/types"
)

// Query atomic swaps flags
const (
	flagInvolve    = "involve"
	flagExpiration = "expiration"
	flagStatus     = "status"
	flagDirection  = "direction"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group bep3 queries under a subcommand
	bep3QueryCmd := &cobra.Command{
		Use:                        "bep3",
		Short:                      "Querying commands for the bep3 module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	bep3QueryCmd.AddCommand(flags.GetCommands(
		QueryCalcSwapIDCmd(queryRoute, cdc),
		QueryCalcRandomNumberHashCmd(queryRoute, cdc),
		QueryGetAssetSupplyCmd(queryRoute, cdc),
		QueryGetAssetSuppliesCmd(queryRoute, cdc),
		QueryGetAtomicSwapCmd(queryRoute, cdc),
		QueryGetAtomicSwapsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
	)...)

	return bep3QueryCmd
}

// QueryCalcRandomNumberHashCmd calculates the random number hash for a number and timestamp
func QueryCalcRandomNumberHashCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "calc-rnh [unix-timestamp]",
		Short:   "calculates an example random number hash from an optional timestamp",
		Example: "bep3 calc-rnh now",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			userTimestamp := "now"
			if len(args) > 0 {
				userTimestamp = args[0]
			}

			// Timestamp defaults to time.Now() unless it's explicitly set
			var timestamp int64
			if strings.Compare(userTimestamp, "now") == 0 {
				timestamp = tmtime.Now().Unix()
			} else {
				userTimestamp, err := strconv.ParseInt(userTimestamp, 10, 64)
				if err != nil {
					return err
				}
				timestamp = userTimestamp
			}

			// Load hex-encoded cryptographically strong pseudo-random number
			randomNumber, err := types.GenerateSecureRandomNumber()
			if err != nil {
				return err
			}
			randomNumberHash := types.CalculateRandomHash(randomNumber, timestamp)

			// Prepare random number, timestamp, and hash for output
			randomNumberStr := fmt.Sprintf("Random number: %s\n", string(randomNumber))
			timestampStr := fmt.Sprintf("Timestamp: %d\n", timestamp)
			randomNumberHashStr := fmt.Sprintf("Random number hash: %s", hex.EncodeToString(randomNumberHash))
			output := []string{randomNumberStr, timestampStr, randomNumberHashStr}
			return cliCtx.PrintOutput(strings.Join(output, ""))
		},
	}
}

// QueryCalcSwapIDCmd calculates the swapID for a random number hash, sender, and sender other chain
func QueryCalcSwapIDCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "calc-swapid [random-number-hash] [sender] [sender-other-chain]",
		Short:   "calculate swap ID for the given random number hash, sender, and sender other chain",
		Example: "bep3 calc-swapid 0677bd8a303dd981810f34d8e5cc6507f13b391899b84d3c1be6c6045a17d747 kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny bnb1ud3q90r98l3mhd87kswv3h8cgrymzeljct8qn7",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Parse query params
			randomNumberHash, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			sender := sdk.AccAddress(args[1])
			senderOtherChain := args[2]

			// Calculate swap ID and convert to human-readable string
			swapID := types.CalculateSwapID(randomNumberHash, sender, senderOtherChain)
			return cliCtx.PrintOutput(hex.EncodeToString(swapID))
		},
	}
}

// QueryGetAssetSupplyCmd queries as asset's current in swap supply, active, supply, and supply limit
func QueryGetAssetSupplyCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "supply [denom]",
		Short:   "get information about an asset's supply",
		Example: "bep3 supply bnb",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare query params
			bz, err := cdc.MarshalJSON(types.NewQueryAssetSupply([]byte(args[0])))
			if err != nil {
				return err
			}

			// Execute query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAssetSupply), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var assetSupply types.AssetSupply
			cdc.MustUnmarshalJSON(res, &assetSupply)
			return cliCtx.PrintOutput(assetSupply)
		},
	}
}

// QueryGetAssetSuppliesCmd queries AssetSupplies in the store
func QueryGetAssetSuppliesCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "supplies",
		Short:   "get a list of all asset supplies",
		Example: "bep3 supplies",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAssetSupplies), nil)
			if err != nil {
				return err
			}

			var assetSupplies types.AssetSupplies
			cdc.MustUnmarshalJSON(res, &assetSupplies)

			if len(assetSupplies) == 0 {
				return fmt.Errorf("There are currently no asset supplies")
			}

			cliCtx = cliCtx.WithHeight(height)
			return cliCtx.PrintOutput(assetSupplies)
		},
	}
}

// QueryGetAtomicSwapCmd queries an AtomicSwap by swapID
func QueryGetAtomicSwapCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "swap [swap-id]",
		Short:   "get atomic swap information",
		Example: "bep3 swap 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Decode swapID's hex encoded string to []byte
			swapID, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			// Prepare query params
			bz, err := cdc.MarshalJSON(types.NewQueryAtomicSwapByID(swapID))
			if err != nil {
				return err
			}

			// Execute query
			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAtomicSwap), bz)
			if err != nil {
				return err
			}

			var atomicSwap types.AtomicSwap
			cdc.MustUnmarshalJSON(res, &atomicSwap)

			cliCtx = cliCtx.WithHeight(height)
			return cliCtx.PrintOutput(atomicSwap.String())
		},
	}
}

// QueryGetAtomicSwapsCmd queries AtomicSwaps in the store
func QueryGetAtomicSwapsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swaps",
		Short: "query atomic swaps with optional filters",
		Long: strings.TrimSpace(`Query for all paginated atomic swaps that match optional filters:
Example:
$ kvcli q bep3 swaps --involve=kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
$ kvcli q bep3 swaps --expiration=280
$ kvcli q bep3 swaps --status=(Open|Completed|Expired)
$ kvcli q bep3 swaps --direction=(Incoming|Outgoing)
$ kvcli q bep3 swaps --page=2 --limit=100
`,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			bechInvolveAddr := viper.GetString(flagInvolve)
			strExpiration := viper.GetString(flagExpiration)
			strSwapStatus := viper.GetString(flagStatus)
			strSwapDirection := viper.GetString(flagDirection)
			page := viper.GetInt(flags.FlagPage)
			limit := viper.GetInt(flags.FlagLimit)

			var involveAddr sdk.AccAddress
			var expiration uint64
			var swapStatus types.SwapStatus
			var swapDirection types.SwapDirection

			params := types.NewQueryAtomicSwaps(page, limit, involveAddr, expiration, swapStatus, swapDirection)

			if len(bechInvolveAddr) != 0 {
				involveAddr, err := sdk.AccAddressFromBech32(bechInvolveAddr)
				if err != nil {
					return err
				}
				params.Involve = involveAddr
			}

			if len(strExpiration) != 0 {
				expiration, err := strconv.ParseUint(strExpiration, 10, 64)
				if err != nil {
					return err
				}
				params.Expiration = expiration
			}

			if len(strSwapStatus) != 0 {
				swapStatus := types.NewSwapStatusFromString(strSwapStatus)
				if !swapStatus.IsValid() {
					return fmt.Errorf("invalid swap status %s", strSwapStatus)
				}
				params.Status = swapStatus
			}

			if len(strSwapDirection) != 0 {
				swapDirection := types.NewSwapDirectionFromString(strSwapDirection)
				if !swapDirection.IsValid() {
					return fmt.Errorf("invalid swap direction %s", strSwapDirection)
				}
				params.Direction = swapDirection
			}

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAtomicSwaps), bz)
			if err != nil {
				return err
			}

			var matchingAtomicSwaps types.AtomicSwaps
			cdc.UnmarshalJSON(res, &matchingAtomicSwaps)

			if len(matchingAtomicSwaps) == 0 {
				return fmt.Errorf("No matching atomic swaps found")
			}

			cliCtx = cliCtx.WithHeight(height)
			return cliCtx.PrintOutput(matchingAtomicSwaps.String()) // nolint:errcheck
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of atomic swaps to to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of atomic swaps to query for")
	cmd.Flags().String(flagInvolve, "", "(optional) filter by atomic swaps that involve an address")
	cmd.Flags().String(flagExpiration, "", "(optional) filter by atomic swaps that expire before a block height")
	cmd.Flags().String(flagStatus, "", "(optional) filter by atomic swap status, status: open/completed/expired")
	cmd.Flags().String(flagDirection, "", "(optional) filter by atomic swap direction, direction: incoming/outgoing")

	return cmd
}

// QueryParamsCmd queries the bep3 module parameters
func QueryParamsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "params",
		Short:   "get the bep3 module parameters",
		Example: "bep3 params",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetParams)
			res, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			// Decode and print results
			var out types.Params
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
