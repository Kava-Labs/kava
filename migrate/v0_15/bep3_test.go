package v0_15

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/x/bep3"
)

var (
	exampleExportTime = time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	exampleBep3Params = bep3.NewParams(bep3.AssetParams{
		bep3.NewAssetParam(
			"xrpb",
			144,
			bep3.SupplyLimit{},
			true,
			sdk.AccAddress("testAddress"),
			sdk.NewInt(1e4),
			sdk.NewInt(1e4),
			sdk.NewInt(1e14),
			24686,
			86400,
		),
	})
	exampleAssetSupplies = bep3.AssetSupplies{
		bep3.NewAssetSupply(
			sdk.NewInt64Coin("xrpb", 1e10),
			sdk.NewInt64Coin("xrpb", 1e9),
			sdk.NewInt64Coin("xrpb", 1e15),
			sdk.NewInt64Coin("xrpb", 0),
			0,
		),
	}
	exampleBep3GenState = bep3.NewGenesisState(
		exampleBep3Params,
		bep3.AtomicSwaps{},
		exampleAssetSupplies,
		exampleExportTime,
	)
)

func exampleBep3Swap(expireHeight uint64, closeHeight int64, status bep3.SwapStatus) bep3.AtomicSwap {
	return bep3.NewAtomicSwap(
		sdk.NewCoins(sdk.NewInt64Coin("xrpb", 1e10)),
		[]byte("random number hash"),
		expireHeight,
		exampleExportTime.Unix(),
		sdk.AccAddress("sender address"),
		sdk.AccAddress("recipient address"),
		"sender other chain address",
		"recipient other chain address",
		closeHeight,
		status,
		true,
		bep3.Outgoing,
	)
}

func TestBep3_SwapHeightsAreReset(t *testing.T) {

	oldState := bep3.NewGenesisState(
		exampleBep3Params,
		bep3.AtomicSwaps{
			exampleBep3Swap(7e5, 6e5, bep3.Open),
			exampleBep3Swap(4e5, 3e5, bep3.Expired),
			exampleBep3Swap(2e5, 1e5, bep3.Completed),
		},
		exampleAssetSupplies,
		exampleExportTime,
	)

	newState := Bep3(oldState)

	expectedSwaps := bep3.AtomicSwaps{
		exampleBep3Swap(1, 6e5, bep3.Expired),
		exampleBep3Swap(1, 3e5, bep3.Expired),
		exampleBep3Swap(2e5, 1, bep3.Completed),
	}

	require.Equal(t, expectedSwaps, newState.AtomicSwaps)
}

func TestBep3_OnlySwapHeightsModified(t *testing.T) {

	oldState := bep3.NewGenesisState(
		exampleBep3Params,
		nil,
		exampleAssetSupplies,
		exampleExportTime,
	)

	newState := Bep3(oldState)

	require.Equal(t, oldState, newState)
}
