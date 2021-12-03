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
		newSwaps[i] = swap
	}
	return newSwaps
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
