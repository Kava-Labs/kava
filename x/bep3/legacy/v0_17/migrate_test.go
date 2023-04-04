package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/bytes"

	app "github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses   []sdk.AccAddress
	v16genstate types.GenesisState
	cdc         codec.Codec
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v16genstate = types.GenesisState{
		PreviousBlockTime: time.Date(2021, 4, 8, 15, 0, 0, 0, time.UTC),
		Params:            types.Params{},
		Supplies:          types.AssetSupplies{},
		AtomicSwaps:       types.AtomicSwaps{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	// Migrate v16 bep3 to v17
	file := filepath.Join("testdata", "v16-bep3.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	err = s.cdc.UnmarshalJSON(data, &s.v16genstate)
	s.Require().NoError(err)
	genstate := Migrate(s.v16genstate)

	// Compare expect v16 bep3 json with migrated json
	actual := s.cdc.MustMarshalJSON(genstate)
	file = filepath.Join("testdata", "v17-bep3.json")
	expected, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_Swaps() {
	type swap struct {
		ExpireHeight uint64
		CloseBlock   int64
		Status       types.SwapStatus
		Direction    types.SwapDirection
	}
	testcases := []struct {
		name    string
		oldSwap swap
		newSwap swap
	}{
		{
			name: "incoming open swap",
			oldSwap: swap{
				// expire and close not set in open swaps
				Status:    types.SWAP_STATUS_OPEN,
				Direction: types.SWAP_DIRECTION_INCOMING,
			},
			newSwap: swap{
				ExpireHeight: 1,
				Status:       types.SWAP_STATUS_EXPIRED,
				Direction:    types.SWAP_DIRECTION_INCOMING,
			},
		},
		{
			name: "outgoing open swap",
			oldSwap: swap{
				// expire and close not set in open swaps
				Status:    types.SWAP_STATUS_OPEN,
				Direction: types.SWAP_DIRECTION_OUTGOING,
			},
			newSwap: swap{
				ExpireHeight: 24687,
				Status:       types.SWAP_STATUS_OPEN,
				Direction:    types.SWAP_DIRECTION_OUTGOING,
			},
		},
		{
			name: "completed swap",
			oldSwap: swap{
				ExpireHeight: 1000,
				CloseBlock:   900,
				Status:       types.SWAP_STATUS_COMPLETED,
				Direction:    types.SWAP_DIRECTION_INCOMING,
			},
			newSwap: swap{
				ExpireHeight: 1000,
				CloseBlock:   1,
				Status:       types.SWAP_STATUS_COMPLETED,
				Direction:    types.SWAP_DIRECTION_INCOMING,
			},
		},
		{
			name: "expired swap",
			oldSwap: swap{
				ExpireHeight: 1000,
				CloseBlock:   900,
				Status:       types.SWAP_STATUS_EXPIRED,
				Direction:    types.SWAP_DIRECTION_INCOMING,
			},
			newSwap: swap{
				ExpireHeight: 1,
				CloseBlock:   900,
				Status:       types.SWAP_STATUS_EXPIRED,
				Direction:    types.SWAP_DIRECTION_INCOMING,
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			oldSwaps := types.AtomicSwaps{
				{
					Amount:              sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(12))),
					RandomNumberHash:    bytes.HexBytes{},
					ExpireHeight:        tc.oldSwap.ExpireHeight,
					Timestamp:           1110,
					Sender:              s.addresses[0],
					Recipient:           s.addresses[1],
					RecipientOtherChain: s.addresses[0].String(),
					SenderOtherChain:    s.addresses[1].String(),
					ClosedBlock:         tc.oldSwap.CloseBlock,
					Status:              tc.oldSwap.Status,
					CrossChain:          true,
					Direction:           tc.oldSwap.Direction,
				},
			}
			expectedSwaps := types.AtomicSwaps{
				{
					Amount:              sdk.NewCoins(sdk.NewCoin("bnb", sdkmath.NewInt(12))),
					RandomNumberHash:    bytes.HexBytes{},
					ExpireHeight:        tc.newSwap.ExpireHeight,
					Timestamp:           1110,
					Sender:              s.addresses[0],
					Recipient:           s.addresses[1],
					RecipientOtherChain: s.addresses[0].String(),
					SenderOtherChain:    s.addresses[1].String(),
					ClosedBlock:         tc.newSwap.CloseBlock,
					Status:              tc.newSwap.Status,
					CrossChain:          true,
					Direction:           tc.newSwap.Direction,
				},
			}
			s.v16genstate.AtomicSwaps = oldSwaps
			genState := Migrate(s.v16genstate)
			s.Require().Len(genState.AtomicSwaps, 1)
			s.Equal(expectedSwaps, genState.AtomicSwaps)
		})
	}
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
