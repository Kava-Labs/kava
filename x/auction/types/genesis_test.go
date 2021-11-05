package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testCoin = sdk.NewInt64Coin("test", 20)

func TestGenesisState_Validate(t *testing.T) {

	defaultGenesisAuctions, err := UnpackGenesisAuctions(DefaultGenesisState().Auctions)
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
			DefaultGenesisState().NextAuctionId,
			defaultGenesisAuctions,
			true,
		},
		{
			"invalid next ID",
			54,
			[]GenesisAuction{
				GenesisAuction(&SurplusAuction{BaseAuction{Id: 105}}),
			},
			false,
		},
		{
			"repeated ID",
			1000,
			[]GenesisAuction{
				GenesisAuction(&SurplusAuction{BaseAuction{Id: 105}}),
				GenesisAuction(&DebtAuction{BaseAuction{Id: 106}, testCoin}),
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gs := NewGenesisState(tc.nextID, DefaultParams(), tc.auctions)

			err := gs.Validate()

			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}

}
