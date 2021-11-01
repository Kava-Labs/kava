package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
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
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group bep3 queries under a subcommand
	bep3QueryCmd := &cobra.Command{
		Use:                        "bep3",
		Short:                      "Querying commands for the bep3 module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmds := []*cobra.Command{
		QueryCalcSwapIDCmd(queryRoute),
		QueryCalcRandomNumberHashCmd(queryRoute),
		QueryGetAssetSupplyCmd(queryRoute),
		QueryGetAssetSuppliesCmd(queryRoute),
		QueryGetAtomicSwapCmd(queryRoute),
		QueryGetAtomicSwapsCmd(queryRoute),
		QueryParamsCmd(queryRoute),
	}

	for _, cmd := range cmds {
		flags.AddQueryFlagsToCmd(cmd)
	}

	bep3QueryCmd.AddCommand(cmds...)

	return bep3QueryCmd
}

// QueryCalcRandomNumberHashCmd calculates the random number hash for a number and timestamp
func QueryCalcRandomNumberHashCmd(queryRoute string) *cobra.Command {
	return &cobra.Command{
		Use:     "calc-rnh [unix-timestamp]",
		Short:   "calculates an example random number hash from an optional timestamp",
		Example: "bep3 calc-rnh now",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

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
			randomNumberStr := fmt.Sprintf("Random number: %s\n", hex.EncodeToString(randomNumber))
			timestampStr := fmt.Sprintf("Timestamp: %d\n", timestamp)
			randomNumberHashStr := fmt.Sprintf("Random number hash: %s", hex.EncodeToString(randomNumberHash))
			output := []string{randomNumberStr, timestampStr, randomNumberHashStr}
			return clientCtx.PrintObjectLegacy(strings.Join(output, ""))
		},
	}
}

// QueryCalcSwapIDCmd calculates the swapID for a random number hash, sender, and sender other chain
func QueryCalcSwapIDCmd(queryRoute string) *cobra.Command {
	return &cobra.Command{
		Use:     "calc-swapid [random-number-hash] [sender] [sender-other-chain]",
		Short:   "calculate swap ID for the given random number hash, sender, and sender other chain",
		Example: "bep3 calc-swapid 0677bd8a303dd981810f34d8e5cc6507f13b391899b84d3c1be6c6045a17d747 kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny bnb1ud3q90r98l3mhd87kswv3h8cgrymzeljct8qn7",
		Args:    cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Parse query params
			randomNumberHash, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			sender, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			senderOtherChain := args[2]

			// Calculate swap ID and convert to human-readable string
			swapID := types.CalculateSwapID(randomNumberHash, sender.String(), senderOtherChain)
			return clientCtx.PrintObjectLegacy(hex.EncodeToString(swapID))
		},
	}
}

// QueryGetAssetSupplyCmd queries as asset's current in swap supply, active, supply, and supply limit
func QueryGetAssetSupplyCmd(queryRoute string) *cobra.Command {
	return &cobra.Command{
		Use:     "supply [denom]",
		Short:   "get information about an asset's supply",
		Example: "bep3 supply bnb",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AssetSupply(context.Background(), &types.QueryAssetSupplyRequest{
				Denom: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

// QueryGetAssetSuppliesCmd queries AssetSupplies in the store
func QueryGetAssetSuppliesCmd(queryRoute string) *cobra.Command {
	return &cobra.Command{
		Use:     "supplies",
		Short:   "get a list of all asset supplies",
		Example: "bep3 supplies",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AssetSupplies(context.Background(), &types.QueryAssetSuppliesRequest{
				// TODO: Pagination here?
			})
			if err != nil {
				return err
			}

			if len(res.AssetSupplies) == 0 {
				return fmt.Errorf("there are currently no asset supplies")
			}

			return clientCtx.PrintProto(res)
		},
	}
}

// QueryGetAtomicSwapCmd queries an AtomicSwap by swapID
func QueryGetAtomicSwapCmd(queryRoute string) *cobra.Command {
	return &cobra.Command{
		Use:     "swap [swap-id]",
		Short:   "get atomic swap information",
		Example: "bep3 swap 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// Decode swapID's hex encoded string to []byte
			swapID, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.AtomicSwap(context.Background(), &types.QueryAtomicSwapRequest{
				SwapId: swapID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
}

// QueryGetAtomicSwapsCmd queries AtomicSwaps in the store
func QueryGetAtomicSwapsCmd(queryRoute string) *cobra.Command {
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
			bechInvolveAddr, err := cmd.Flags().GetString(flagInvolve)
			if err != nil {
				return err
			}
			strExpiration, err := cmd.Flags().GetString(flagExpiration)
			if err != nil {
				return err
			}
			strSwapStatus, err := cmd.Flags().GetString(flagStatus)
			if err != nil {
				return err
			}
			strSwapDirection, err := cmd.Flags().GetString(flagDirection)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := types.QueryAtomicSwapsRequest{
				Pagination: pageReq,
			}

			if len(bechInvolveAddr) != 0 {
				involveAddr, err := sdk.AccAddressFromBech32(bechInvolveAddr)
				if err != nil {
					return err
				}
				req.Involve = involveAddr.String()
			}

			if len(strExpiration) != 0 {
				expiration, err := strconv.ParseUint(strExpiration, 10, 64)
				if err != nil {
					return err
				}
				req.Expiration = expiration
			}

			if len(strSwapStatus) != 0 {
				swapStatus := types.NewSwapStatusFromString(strSwapStatus)
				if !swapStatus.IsValid() {
					return fmt.Errorf("invalid swap status %s", strSwapStatus)
				}
				req.Status = swapStatus
			}

			if len(strSwapDirection) != 0 {
				swapDirection := types.NewSwapDirectionFromString(strSwapDirection)
				if !swapDirection.IsValid() {
					return fmt.Errorf("invalid swap direction %s", strSwapDirection)
				}
				req.Direction = swapDirection
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AtomicSwaps(context.Background(), &req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagInvolve, "", "(optional) filter by atomic swaps that involve an address")
	cmd.Flags().String(flagExpiration, "", "(optional) filter by atomic swaps that expire before a block height")
	cmd.Flags().String(flagStatus, "", "(optional) filter by atomic swap status, status: open/completed/expired")
	cmd.Flags().String(flagDirection, "", "(optional) filter by atomic swap direction, direction: incoming/outgoing")

	flags.AddPaginationFlagsToCmd(cmd, "swaps")

	return cmd
}

// QueryParamsCmd queries the bep3 module parameters
func QueryParamsCmd(queryRoute string) *cobra.Command {
	return &cobra.Command{
		Use:     "params",
		Short:   "get the bep3 module parameters",
		Example: "bep3 params",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
}
