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
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/x/bep3/types"
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
		QueryGetAtomicSwapCmd(queryRoute, cdc),
		QueryGetAssetSupplyCmd(queryRoute, cdc),
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
			randomNumberStr := fmt.Sprintf("Random number: %s\n", randomNumber)
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
		Example: "bep3 calc-swapid 0677bd8a303dd981810f34d8e5cc6507f13b391899b84d3c1be6c6045a17d747 kava15qdefkmwswysgg4qxgcqpqr35k3m49pkx2jdfnw bnb1ud3q90r98l3mhd87kswv3h8cgrymzeljct8qn7",
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
	return &cobra.Command{
		Use:     "swaps",
		Short:   "get a list of active atomic swaps",
		Example: "bep3 swaps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetAtomicSwaps), nil)
			if err != nil {
				return err
			}

			var atomicSwaps types.AtomicSwaps
			cdc.MustUnmarshalJSON(res, &atomicSwaps)

			if len(atomicSwaps) == 0 {
				return fmt.Errorf("There are currently no atomic swaps")
			}

			cliCtx = cliCtx.WithHeight(height)
			return cliCtx.PrintOutput(atomicSwaps.String())
		},
	}
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
