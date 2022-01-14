package v0_16

import (
	"fmt"

	v015bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_15"
	v016bep3 "github.com/kava-labs/kava/x/bep3/types"
)

func migrateAssetParam(param v015bep3.AssetParam) v016bep3.AssetParam {
	return v016bep3.AssetParam{
		Denom:  param.Denom,
		CoinID: int64(param.CoinID),
		SupplyLimit: v016bep3.SupplyLimit{
			Limit:          param.SupplyLimit.Limit,
			TimeLimited:    param.SupplyLimit.TimeLimited,
			TimePeriod:     param.SupplyLimit.TimePeriod,
			TimeBasedLimit: param.SupplyLimit.TimeBasedLimit,
		},
		Active:        param.Active,
		DeputyAddress: param.DeputyAddress,
		FixedFee:      param.FixedFee,
		MinSwapAmount: param.MinSwapAmount,
		MaxSwapAmount: param.MaxSwapAmount,
		MinBlockLock:  param.MinBlockLock,
		MaxBlockLock:  param.MaxBlockLock,
	}
}

func migrateParams(v015params v015bep3.Params) v016bep3.Params {
	assetParams := make(v016bep3.AssetParams, len(v015params.AssetParams))
	for i, assetParam := range v015params.AssetParams {
		assetParams[i] = migrateAssetParam(assetParam)
	}
	return v016bep3.Params{AssetParams: assetParams}
}

func migrateSwapStatus(status v015bep3.SwapStatus) v016bep3.SwapStatus {
	switch status {
	case v015bep3.NULL:
		return v016bep3.SWAP_STATUS_UNSPECIFIED
	case v015bep3.Open:
		return v016bep3.SWAP_STATUS_OPEN
	case v015bep3.Completed:
		return v016bep3.SWAP_STATUS_COMPLETED
	case v015bep3.Expired:
		return v016bep3.SWAP_STATUS_EXPIRED
	default:
		panic(fmt.Errorf("'%d' is not a valid swap status", status))
	}
}

func migrateSwapDirection(direction v015bep3.SwapDirection) v016bep3.SwapDirection {
	switch direction {
	case v015bep3.INVALID:
		return v016bep3.SWAP_DIRECTION_UNSPECIFIED
	case v015bep3.Incoming:
		return v016bep3.SWAP_DIRECTION_INCOMING
	case v015bep3.Outgoing:
		return v016bep3.SWAP_DIRECTION_OUTGOING
	default:
		panic(fmt.Errorf("'%d' is not a valid swap direction", direction))
	}
}

func migrateAtomicSwaps(oldSwaps v015bep3.AtomicSwaps) v016bep3.AtomicSwaps {
	newSwaps := make(v016bep3.AtomicSwaps, len(oldSwaps))
	for i, oldSwap := range oldSwaps {
		swap := v016bep3.AtomicSwap{
			Amount:              oldSwap.Amount,
			RandomNumberHash:    oldSwap.RandomNumberHash,
			ExpireHeight:        oldSwap.ExpireHeight,
			Timestamp:           oldSwap.Timestamp,
			Sender:              oldSwap.Sender,
			Recipient:           oldSwap.Recipient,
			SenderOtherChain:    oldSwap.SenderOtherChain,
			RecipientOtherChain: oldSwap.RecipientOtherChain,
			ClosedBlock:         oldSwap.ClosedBlock,
			Status:              migrateSwapStatus(oldSwap.Status),
			CrossChain:          oldSwap.CrossChain,
			Direction:           migrateSwapDirection(oldSwap.Direction),
		}
		swap = resetSwapForZeroHeight(swap)
		newSwaps[i] = swap
	}
	return newSwaps
}

// resetSwapForZeroHeight updates swap expiry/close heights to work when the chain height is reset to zero.
func resetSwapForZeroHeight(swap v016bep3.AtomicSwap) v016bep3.AtomicSwap {
	switch status := swap.Status; status {
	case v016bep3.SWAP_STATUS_COMPLETED:
		// Reset closed block to one so completed swaps are not held in long term storage too long.
		swap.ClosedBlock = 1
	case v016bep3.SWAP_STATUS_OPEN:
		switch dir := swap.Direction; dir {
		case v016bep3.SWAP_DIRECTION_INCOMING:
			// Open incoming swaps can be expired safely. They haven't been claimed yet, so the outgoing swap on bnb will just timeout.
			// The chain downtime cannot be accurately predicted, so it's easier to expire than to recalculate a correct expire height.
			swap.ExpireHeight = 1
			swap.Status = v016bep3.SWAP_STATUS_EXPIRED
		case v016bep3.SWAP_DIRECTION_OUTGOING:
			// Open outgoing swaps should be extended to allow enough time to claim after the chain launches.
			// They cannot be expired as there could be an open/claimed bnb swap.
			swap.ExpireHeight = 1 + 24686 // default timeout used when sending swaps from kava
		case v016bep3.SWAP_DIRECTION_UNSPECIFIED:
		default:
			panic(fmt.Sprintf("unknown bep3 swap direction '%s'", dir))
		}
	case v016bep3.SWAP_STATUS_EXPIRED:
		// Once a swap is marked expired the expire height is ignored. However reset to 1 to be sure.
		swap.ExpireHeight = 1
	case v016bep3.SWAP_STATUS_UNSPECIFIED:
	default:
		panic(fmt.Sprintf("unknown bep3 swap status '%s'", status))
	}

	return swap
}

func migrateSupplies(oldSupplies v015bep3.AssetSupplies) v016bep3.AssetSupplies {
	newSupplies := make(v016bep3.AssetSupplies, len(oldSupplies))
	for i, supply := range oldSupplies {
		newSupplies[i] = v016bep3.AssetSupply{
			IncomingSupply:           supply.IncomingSupply,
			OutgoingSupply:           supply.OutgoingSupply,
			CurrentSupply:            supply.CurrentSupply,
			TimeLimitedCurrentSupply: supply.TimeLimitedCurrentSupply,
			TimeElapsed:              supply.TimeElapsed,
		}
	}
	return newSupplies
}

func Migrate(oldState v015bep3.GenesisState) *v016bep3.GenesisState {
	return &v016bep3.GenesisState{
		PreviousBlockTime: oldState.PreviousBlockTime,
		Params:            migrateParams(oldState.Params),
		AtomicSwaps:       migrateAtomicSwaps(oldState.AtomicSwaps),
		Supplies:          migrateSupplies(oldState.Supplies),
	}
}
