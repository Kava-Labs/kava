package v0_16

import (
	"fmt"

	"github.com/kava-labs/kava/x/bep3/types"
)

// resetSwapForZeroHeight updates swap expiry/close heights to work when the chain height is reset to zero.
func resetSwapForZeroHeight(swap types.AtomicSwap) types.AtomicSwap {
	switch status := swap.Status; status {
	case types.SWAP_STATUS_COMPLETED:
		// Reset closed block to one so completed swaps are not held in long term storage too long.
		swap.ClosedBlock = 1
	case types.SWAP_STATUS_OPEN:
		switch dir := swap.Direction; dir {
		case types.SWAP_DIRECTION_INCOMING:
			// Open incoming swaps can be expired safely. They haven't been claimed yet, so the outgoing swap on bnb will just timeout.
			// The chain downtime cannot be accurately predicted, so it's easier to expire than to recalculate a correct expire height.
			swap.ExpireHeight = 1
			swap.Status = types.SWAP_STATUS_EXPIRED
		case types.SWAP_DIRECTION_OUTGOING:
			// Open outgoing swaps should be extended to allow enough time to claim after the chain launches.
			// They cannot be expired as there could be an open/claimed bnb swap.
			swap.ExpireHeight = 1 + 24686 // default timeout used when sending swaps from kava
		case types.SWAP_DIRECTION_UNSPECIFIED:
		default:
			panic(fmt.Sprintf("unknown bep3 swap direction '%s'", dir))
		}
	case types.SWAP_STATUS_EXPIRED:
		// Once a swap is marked expired the expire height is ignored. However reset to 1 to be sure.
		swap.ExpireHeight = 1
	case types.SWAP_STATUS_UNSPECIFIED:
	default:
		panic(fmt.Sprintf("unknown bep3 swap status '%s'", status))
	}

	return swap
}

func resetSwapsForZeroHeight(oldSwaps types.AtomicSwaps) types.AtomicSwaps {
	newSwaps := make(types.AtomicSwaps, len(oldSwaps))
	for i, oldSwap := range oldSwaps {
		swap := resetSwapForZeroHeight(oldSwap)
		newSwaps[i] = swap
	}
	return newSwaps
}

func Migrate(oldState types.GenesisState) *types.GenesisState {
	return &types.GenesisState{
		PreviousBlockTime: oldState.PreviousBlockTime,
		Params:            oldState.Params,
		AtomicSwaps:       resetSwapsForZeroHeight(oldState.AtomicSwaps),
		Supplies:          oldState.Supplies,
	}
}
