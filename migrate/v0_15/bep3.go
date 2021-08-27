package v0_15

import (
	"fmt"

	v0_15bep3 "github.com/kava-labs/kava/x/bep3/types"
)

// Bep3 resets the swap expire/close heights for a chain starting at height 0.
func Bep3(genesisState v0_15bep3.GenesisState) v0_15bep3.GenesisState {

	var newSwaps v0_15bep3.AtomicSwaps
	for _, swap := range genesisState.AtomicSwaps {

		switch status := swap.Status; status {
		case v0_15bep3.Completed:
			// reset closed block to one so completed swaps are removed from long term storage properly
			swap.ClosedBlock = 1

		case v0_15bep3.Open, v0_15bep3.Expired:
			swap.Status = v0_15bep3.Expired // set open swaps to expired so they can be refunded (by tx) after chain start
			swap.ExpireHeight = 1           // set expire on first block as well to be safe

		default:
			panic(fmt.Sprintf("unknown bep3 swap status '%s'", status))
		}

		newSwaps = append(newSwaps, swap)
	}

	genesisState.AtomicSwaps = newSwaps

	return genesisState
}
