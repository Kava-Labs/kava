package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group bep3 queries under a subcommand
	bep3QueryCmd := &cobra.Command{
		Use:   "bep3",
		Short: "Querying commands for the bep3 module",
	}

	bep3QueryCmd.AddCommand(client.GetCommands(
		QueryCalcSwapIDCmd(queryRoute, cdc),
		QueryCalcRandomNumberHashCmd(queryRoute, cdc),
		QueryGetHtltCmd(queryRoute, cdc),
		QueryGetHtltsCmd(queryRoute, cdc),
		QueryParamsCmd(queryRoute, cdc),
	)...)

	return bep3QueryCmd
}

// QueryCalcRandomNumberHashCmd calculates the random number hash for a number and timestamp
func QueryCalcRandomNumberHashCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "calc-rnh [random-number] [timestamp]",
		Short:   "calculate a random number hash for given a number and timestamp",
		Example: "bep3 calc-rnh 15 9988776655",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Parse query params
			if len(strings.TrimSpace(args[0])) == 0 {
				return fmt.Errorf("random-number cannot be empty")
			}
			randomNumber := []byte(args[0])
			timestamp, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("timestamp %s could not be converted to an integer", args[1]))
			}

			// Calculate random number hash and convert to human-readable string
			randomNumberHash := types.CalculateRandomHash(randomNumber, timestamp)
			decodedRandomNumberHash := types.BytesToHexEncodedString(randomNumberHash)

			return cliCtx.PrintOutput(decodedRandomNumberHash)
		},
	}
}

// QueryCalcSwapIDCmd calculates the swapID for a random number hash, sender, and sender other chain
func QueryCalcSwapIDCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "calc-id [randomNumberHash] [sender] [senderOtherChain]",
		Short:   "calculate swap ID for the given random number hash, sender, and sender other chain",
		Example: "bep3 calc-id 0677bd8a303dd981810f34d8e5cc6507f13b391899b84d3c1be6c6045a17d747 kava15qdefkmwswysgg4qxgcqpqr35k3m49pkx2jdfnw bnb1ud3q90r98l3mhd87kswv3h8cgrymzeljct8qn7",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Parse query params
			if len(strings.TrimSpace(args[0])) != types.RandomNumberHashLength {
				return fmt.Errorf("random-number-hash should have length %d", types.RandomNumberHashLength)
			}
			randomNumberHash := args[0]

			sender := sdk.AccAddress(args[1])
			senderOtherChain := args[2]

			bytesRNH, err := types.HexEncodedStringToBytes(randomNumberHash)
			if err != nil {
				return err
			}

			// Calculate swap ID and convert to human-readable string
			swapIDBytes, err := types.CalculateSwapID(bytesRNH, sender, senderOtherChain)
			if err != nil {
				return err
			}
			swapID := types.BytesToHexEncodedString(swapIDBytes)

			return cliCtx.PrintOutput(swapID)
		},
	}
}

// QueryGetHtltCmd queries an HTLT by swapID
func QueryGetHtltCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "htlt [swap-id]",
		Short:   "get HTLT information",
		Example: "bep3 htlt 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Decode swapID's hex encoded string to []byte
			swapID, err := types.HexEncodedStringToBytes(args[0])
			if err != nil {
				return err
			}

			// Prepare query params
			bz, err := cdc.MarshalJSON(types.NewQueryHTLTByID(swapID))
			if err != nil {
				return err
			}

			// Execute query
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetHTLT), bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var htlt types.HTLT
			cdc.MustUnmarshalJSON(res, &htlt)
			return cliCtx.PrintOutput(htlt)
		},
	}
}

// QueryGetHtltsCmd queries the htlts in the store
func QueryGetHtltsCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "htlts",
		Short:   "get a list of active htlts",
		Example: "bep3 htlts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryGetHTLTs), nil)
			if err != nil {
				return err
			}

			var htlts types.HTLTs
			cdc.MustUnmarshalJSON(res, &htlts)

			if len(htlts) == 0 {
				return fmt.Errorf("There are currently no htlts")
			}

			cliCtx = cliCtx.WithHeight(height)
			return cliCtx.PrintOutput(htlts)
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
