package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/bytes"

	app "github.com/kava-labs/kava/app"
	v015bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_15"
	v016bep3 "github.com/kava-labs/kava/x/bep3/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses   []sdk.AccAddress
	v15genstate v015bep3.GenesisState
	cdc         codec.Codec
	legacyCdc   *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v15genstate = v015bep3.GenesisState{
		PreviousBlockTime: time.Date(2021, 4, 8, 15, 0, 0, 0, time.UTC),
		Params:            v015bep3.Params{},
		Supplies:          v015bep3.AssetSupplies{},
		AtomicSwaps:       v015bep3.AtomicSwaps{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	// Migrate v15 bep3 to v16
	file := filepath.Join("testdata", "v15-bep3.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	err = s.legacyCdc.UnmarshalJSON(data, &s.v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(s.v15genstate)

	// Compare expect v16 bep3 json with migrated json
	actual := s.cdc.MustMarshalJSON(genstate)
	file = filepath.Join("testdata", "v16-bep3.json")
	expected, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_Swaps() {
	type oldSwap struct {
		ExpireHeight uint64
		CloseBlock   int64
		Status       v015bep3.SwapStatus
		Direction    v015bep3.SwapDirection
	}
	type newSwap struct {
		ExpireHeight uint64
		CloseBlock   int64
		Status       v016bep3.SwapStatus
		Direction    v016bep3.SwapDirection
	}
	testcases := []struct {
		name    string
		oldSwap oldSwap
		newSwap newSwap
	}{
		{
			name: "invalid swap direction",
			oldSwap: oldSwap{
				Direction: v015bep3.INVALID,
			},
			newSwap: newSwap{
				Direction: v016bep3.SWAP_DIRECTION_UNSPECIFIED,
			},
		},
		{
			name: "invalid swap status",
			oldSwap: oldSwap{
				Status: v015bep3.NULL,
			},
			newSwap: newSwap{
				Status: v016bep3.SWAP_STATUS_UNSPECIFIED,
			},
		},
		{
			name: "incoming open swap",
			oldSwap: oldSwap{
				// expire and close not set in open swaps
				Status:    v015bep3.Open,
				Direction: v015bep3.Incoming,
			},
			newSwap: newSwap{
				ExpireHeight: 1,
				Status:       v016bep3.SWAP_STATUS_EXPIRED,
				Direction:    v016bep3.SWAP_DIRECTION_INCOMING,
			},
		},
		{
			name: "outgoing open swap",
			oldSwap: oldSwap{
				// expire and close not set in open swaps
				Status:    v015bep3.Open,
				Direction: v015bep3.Outgoing,
			},
			newSwap: newSwap{
				ExpireHeight: 24687,
				Status:       v016bep3.SWAP_STATUS_OPEN,
				Direction:    v016bep3.SWAP_DIRECTION_OUTGOING,
			},
		},
		{
			name: "completed swap",
			oldSwap: oldSwap{
				ExpireHeight: 1000,
				CloseBlock:   900,
				Status:       v015bep3.Completed,
				Direction:    v015bep3.Incoming,
			},
			newSwap: newSwap{
				ExpireHeight: 1000,
				CloseBlock:   1,
				Status:       v016bep3.SWAP_STATUS_COMPLETED,
				Direction:    v016bep3.SWAP_DIRECTION_INCOMING,
			},
		},
		{
			name: "expired swap",
			oldSwap: oldSwap{
				ExpireHeight: 1000,
				CloseBlock:   900,
				Status:       v015bep3.Expired,
				Direction:    v015bep3.Incoming,
			},
			newSwap: newSwap{
				ExpireHeight: 1,
				CloseBlock:   900,
				Status:       v016bep3.SWAP_STATUS_EXPIRED,
				Direction:    v016bep3.SWAP_DIRECTION_INCOMING,
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			oldSwaps := v015bep3.AtomicSwaps{
				{
					Amount:              sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(12))),
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
			expectedSwaps := v016bep3.AtomicSwaps{
				{
					Amount:              sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(12))),
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
			s.v15genstate.AtomicSwaps = oldSwaps
			genState := Migrate(s.v15genstate)
			s.Require().Len(genState.AtomicSwaps, 1)
			s.Equal(expectedSwaps, genState.AtomicSwaps)
		})
	}
}

func (s *migrateTestSuite) TestMigrate_Params() {
	params := v015bep3.AssetParams{
		{
			Denom:  "bnb",
			CoinID: int(714),
			SupplyLimit: v015bep3.SupplyLimit{
				Limit:          sdk.NewInt(350000000000000),
				TimeLimited:    false,
				TimeBasedLimit: sdk.ZeroInt(),
				TimePeriod:     time.Hour,
			},
			Active:        true,
			DeputyAddress: s.addresses[0],
			FixedFee:      sdk.NewInt(1000),
			MinSwapAmount: sdk.OneInt(),
			MaxSwapAmount: sdk.NewInt(1000000000000),
			MinBlockLock:  220,
			MaxBlockLock:  770,
		},
	}
	expectedParams := v016bep3.AssetParams{
		{
			Denom:  "bnb",
			CoinID: int64(714),
			SupplyLimit: v016bep3.SupplyLimit{
				Limit:          sdk.NewInt(350000000000000),
				TimeLimited:    false,
				TimeBasedLimit: sdk.ZeroInt(),
				TimePeriod:     time.Hour,
			},
			Active:        true,
			DeputyAddress: s.addresses[0],
			FixedFee:      sdk.NewInt(1000),
			MinSwapAmount: sdk.OneInt(),
			MaxSwapAmount: sdk.NewInt(1000000000000),
			MinBlockLock:  220,
			MaxBlockLock:  770,
		},
	}

	s.v15genstate.Params = v015bep3.Params{AssetParams: params}
	genState := Migrate(s.v15genstate)
	s.Require().Len(genState.Params.AssetParams, 1)
	s.Require().Equal(v016bep3.Params{AssetParams: expectedParams}, genState.Params)
}

func (s *migrateTestSuite) TestMigrate_Supplies() {
	supplies := v015bep3.AssetSupplies{
		{
			IncomingSupply:           sdk.NewInt64Coin("bnb", 1000),
			OutgoingSupply:           sdk.NewInt64Coin("bnb", 1001),
			CurrentSupply:            sdk.NewInt64Coin("bnb", 1002),
			TimeLimitedCurrentSupply: sdk.NewInt64Coin("bnb", 1003),
			TimeElapsed:              time.Hour,
		},
	}
	expectedSupplies := v016bep3.AssetSupplies{
		{
			IncomingSupply:           sdk.NewInt64Coin("bnb", 1000),
			OutgoingSupply:           sdk.NewInt64Coin("bnb", 1001),
			CurrentSupply:            sdk.NewInt64Coin("bnb", 1002),
			TimeLimitedCurrentSupply: sdk.NewInt64Coin("bnb", 1003),
			TimeElapsed:              time.Hour,
		},
	}

	s.v15genstate.Supplies = supplies
	genState := Migrate(s.v15genstate)
	s.Require().Len(genState.Supplies, 1)
	s.Require().Equal(expectedSupplies, genState.Supplies)
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
