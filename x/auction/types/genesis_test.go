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

	defaultGenState := DefaultGenesisState()
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

func TestGenesisState_UnmarshalAnys(t *testing.T) {

	cdc := newTestModuleCodec()

	arbitraryTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	genesis, err := NewGenesisState(
		0,
		NewParams(
			time.Hour*24,
			time.Hour,
			sdk.MustNewDecFromStr("0.05"),
			sdk.MustNewDecFromStr("0.05"),
			sdk.MustNewDecFromStr("0.05"),
		),
		[]GenesisAuction{
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
		},
	)
	require.NoError(t, err)

	bz, err := cdc.Marshal(genesis)
	require.NoError(t, err)

	var unmarshalledGenesis GenesisState
	cdc.Unmarshal(bz, &unmarshalledGenesis)

	// note: require.Equal uses reflect.DeepEqual which has access to unexported fields.
	// This allows it to identify when Any.cachedValue has not been populated correctly.
	require.Equal(t, genesis, &unmarshalledGenesis)
}
