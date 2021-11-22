package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testCoin = sdk.NewInt64Coin("test", 20)

func TestGenesisState_Validate(t *testing.T) {

	defaultGenState, err := DefaultGenesisState()
	require.NoError(t, err)

	defaultGenesisAuctions, err := UnpackGenesisAuctions(defaultGenState.Auctions)
	if err != nil {
		panic(err)
	}

	testCases := []struct {
		name       string
		nextID     uint64
		auctions   []GenesisAuction
		expectPass bool
	}{
		{
			"default",
			defaultGenState.NextAuctionId,
			defaultGenesisAuctions,
			true,
		},
		{
			"invalid next ID",
			54,
			[]GenesisAuction{
				GenesisAuction(&SurplusAuction{BaseAuction{ID: 105}}),
			},
			false,
		},
		{
			"repeated ID",
			1000,
			[]GenesisAuction{
				GenesisAuction(&SurplusAuction{BaseAuction{ID: 105}}),
				GenesisAuction(&DebtAuction{BaseAuction{ID: 106}, testCoin}),
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gs, err := NewGenesisState(tc.nextID, DefaultParams(), tc.auctions)
			require.NoError(t, err)

			err = gs.Validate()
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}

}
