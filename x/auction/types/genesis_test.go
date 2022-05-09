package types

import (
	"testing"
	time "time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testCoin = sdk.NewInt64Coin("test", 20)

func newTestModuleCodec() codec.Codec {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	RegisterInterfaces(interfaceRegistry)
	return codec.NewProtoCodec(interfaceRegistry)
}

func TestGenesisState_Validate(t *testing.T) {
	arbitraryTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	validAuction := &CollateralAuction{
		BaseAuction: BaseAuction{
			ID:              10,
			Initiator:       "seller mod account",
			Lot:             sdk.NewInt64Coin("btc", 1e8),
			Bidder:          sdk.AccAddress("test bidder"),
			Bid:             sdk.NewInt64Coin("usdx", 5),
			HasReceivedBids: true,
			EndTime:         arbitraryTime,
			MaxEndTime:      arbitraryTime.Add(time.Hour),
		},
		CorrespondingDebt: sdk.NewInt64Coin("debt", 1e9),
		MaxBid:            sdk.NewInt64Coin("usdx", 5e4),
		LotReturns: WeightedAddresses{
			Addresses: []sdk.AccAddress{sdk.AccAddress("test return address")},
			Weights:   []sdk.Int{sdk.OneInt()},
		},
	}

	testCases := []struct {
		name       string
		genesis    *GenesisState
		expectPass bool
	}{
		{
			"valid default genesis",
			DefaultGenesisState(),
			true,
		},
		{
			"invalid next ID",
			&GenesisState{
				validAuction.ID - 1,
				DefaultParams(),
				mustPackGenesisAuctions(
					[]GenesisAuction{
						validAuction,
					},
				),
			},
			false,
		},
		{
			"invalid auctions with repeated ID",
			&GenesisState{
				validAuction.ID + 1,
				DefaultParams(),
				mustPackGenesisAuctions(
					[]GenesisAuction{
						validAuction,
						validAuction,
					},
				),
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGenesisState_UnmarshalAnys(t *testing.T) {
	cdc := newTestModuleCodec()

	arbitraryTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	auctions := []GenesisAuction{
		&CollateralAuction{
			BaseAuction: BaseAuction{
				ID:              1,
				Initiator:       "seller mod account",
				Lot:             sdk.NewInt64Coin("btc", 1e8),
				Bidder:          sdk.AccAddress("test bidder"),
				Bid:             sdk.NewInt64Coin("usdx", 5),
				HasReceivedBids: true,
				EndTime:         arbitraryTime,
				MaxEndTime:      arbitraryTime.Add(time.Hour),
			},
			CorrespondingDebt: sdk.NewInt64Coin("debt", 1e9),
			MaxBid:            sdk.NewInt64Coin("usdx", 5e4),
			LotReturns:        WeightedAddresses{},
		},
		&DebtAuction{
			BaseAuction: BaseAuction{
				ID:              2,
				Initiator:       "mod account",
				Lot:             sdk.NewInt64Coin("ukava", 1e9),
				Bidder:          sdk.AccAddress("test bidder"),
				Bid:             sdk.NewInt64Coin("usdx", 5),
				HasReceivedBids: true,
				EndTime:         arbitraryTime,
				MaxEndTime:      arbitraryTime.Add(time.Hour),
			},
			CorrespondingDebt: testCoin,
		},
		&SurplusAuction{
			BaseAuction: BaseAuction{
				ID:              3,
				Initiator:       "seller mod account",
				Lot:             sdk.NewInt64Coin("usdx", 1e9),
				Bidder:          sdk.AccAddress("test bidder"),
				Bid:             sdk.NewInt64Coin("ukava", 5),
				HasReceivedBids: true,
				EndTime:         arbitraryTime,
				MaxEndTime:      arbitraryTime.Add(time.Hour),
			},
		},
	}
	genesis, err := NewGenesisState(
		DefaultNextAuctionID,
		DefaultParams(),
		auctions,
	)
	require.NoError(t, err)

	bz, err := cdc.MarshalJSON(genesis)
	require.NoError(t, err)

	var unmarshalledGenesis GenesisState
	cdc.UnmarshalJSON(bz, &unmarshalledGenesis)

	// Check the interface values are correct after unmarshalling.
	unmarshalledAuctions, err := UnpackGenesisAuctions(unmarshalledGenesis.Auctions)
	require.NoError(t, err)
	require.Equal(t, auctions, unmarshalledAuctions)
}
