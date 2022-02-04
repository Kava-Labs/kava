package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	app "github.com/kava-labs/kava/app"
	v015auction "github.com/kava-labs/kava/x/auction/legacy/v0_15"
	v016auction "github.com/kava-labs/kava/x/auction/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses   []sdk.AccAddress
	v15genstate v015auction.GenesisState
	cdc         codec.Codec
	legacyCdc   *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v15genstate = v015auction.GenesisState{
		NextAuctionID: 1,
		Params:        v015auction.Params{},
		Auctions:      v015auction.GenesisAuctions{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	v015auction.RegisterLegacyAminoCodec(legacyCodec)
	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	// Migrate v15 auction to v16
	file := filepath.Join("testdata", "v15-auction.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	err = s.legacyCdc.UnmarshalJSON(data, &s.v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(s.v15genstate)

	// Compare expect v16 auction json with migrated json
	actual := s.cdc.MustMarshalJSON(genstate)
	file = filepath.Join("testdata", "v16-auction.json")
	expected, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_Auction() {
	now := time.Now()
	testcases := []struct {
		name       string
		oldAuction v015auction.GenesisAuction
		newAuction v016auction.GenesisAuction
	}{
		{
			name: "collateral auction",
			oldAuction: v015auction.CollateralAuction{
				BaseAuction: v015auction.BaseAuction{
					ID:              1,
					Initiator:       s.addresses[0].String(),
					Lot:             sdk.NewInt64Coin("kava", 1),
					Bidder:          s.addresses[1],
					Bid:             sdk.NewInt64Coin("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: sdk.NewInt64Coin("kava", 1),
				MaxBid:            sdk.NewInt64Coin("kava", 1),
				LotReturns: v015auction.WeightedAddresses{
					Addresses: s.addresses[:2],
					Weights:   []sdk.Int{sdk.NewInt(1), sdk.NewInt(1)},
				},
			},
			newAuction: &v016auction.CollateralAuction{
				BaseAuction: v016auction.BaseAuction{
					ID:              1,
					Initiator:       s.addresses[0].String(),
					Lot:             sdk.NewInt64Coin("kava", 1),
					Bidder:          s.addresses[1],
					Bid:             sdk.NewInt64Coin("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: sdk.NewInt64Coin("kava", 1),
				MaxBid:            sdk.NewInt64Coin("kava", 1),
				LotReturns: v016auction.WeightedAddresses{
					Addresses: s.addresses[:2],
					Weights:   []sdk.Int{sdk.NewInt(1), sdk.NewInt(1)},
				},
			},
		},
		{
			name: "surplus auction",
			oldAuction: v015auction.SurplusAuction{
				BaseAuction: v015auction.BaseAuction{
					ID:              2,
					Initiator:       s.addresses[0].String(),
					Lot:             sdk.NewInt64Coin("kava", 12),
					Bidder:          s.addresses[1],
					Bid:             sdk.NewInt64Coin("kava", 12),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: false,
				},
			},
			newAuction: &v016auction.SurplusAuction{
				BaseAuction: v016auction.BaseAuction{
					ID:              2,
					Initiator:       s.addresses[0].String(),
					Lot:             sdk.NewInt64Coin("kava", 12),
					Bidder:          s.addresses[1],
					Bid:             sdk.NewInt64Coin("kava", 12),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: false,
				},
			},
		},
		{
			name: "debt auction",
			oldAuction: v015auction.DebtAuction{
				BaseAuction: v015auction.BaseAuction{
					ID:              3,
					Initiator:       s.addresses[0].String(),
					Lot:             sdk.NewInt64Coin("kava", 1),
					Bidder:          s.addresses[1],
					Bid:             sdk.NewInt64Coin("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: sdk.NewInt64Coin("kava", 20),
			},
			newAuction: &v016auction.DebtAuction{
				BaseAuction: v016auction.BaseAuction{
					ID:              3,
					Initiator:       s.addresses[0].String(),
					Lot:             sdk.NewInt64Coin("kava", 1),
					Bidder:          s.addresses[1],
					Bid:             sdk.NewInt64Coin("kava", 1),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: true,
				},
				CorrespondingDebt: sdk.NewInt64Coin("kava", 20),
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.v15genstate.Auctions = v015auction.GenesisAuctions{tc.oldAuction}
			genState := Migrate(s.v15genstate)
			s.Require().Len(genState.Auctions, 1)
			expectedAuctions, err := v016auction.UnpackGenesisAuctions(genState.Auctions)
			s.Require().NoError(err)
			s.Equal(tc.newAuction, expectedAuctions[0])
		})
	}
}

func (s *migrateTestSuite) TestMigrate_GenState() {
	now := time.Now()
	v15params := v015auction.Params{
		MaxAuctionDuration:  time.Duration(time.Hour * 24 * 7),
		BidDuration:         time.Duration(time.Hour * 24 * 4),
		IncrementSurplus:    sdk.MustNewDecFromStr("0.01"),
		IncrementDebt:       sdk.MustNewDecFromStr("0.02"),
		IncrementCollateral: sdk.MustNewDecFromStr("0.03"),
	}
	v16params := v016auction.Params{
		MaxAuctionDuration:  v15params.MaxAuctionDuration,
		BidDuration:         v15params.BidDuration,
		IncrementSurplus:    v15params.IncrementSurplus,
		IncrementDebt:       v15params.IncrementDebt,
		IncrementCollateral: v15params.IncrementCollateral,
	}
	expectedAuction := &v016auction.SurplusAuction{
		BaseAuction: v016auction.BaseAuction{
			ID:              2,
			Initiator:       s.addresses[0].String(),
			Lot:             sdk.NewInt64Coin("kava", 12),
			Bidder:          s.addresses[1],
			Bid:             sdk.NewInt64Coin("kava", 12),
			EndTime:         now,
			MaxEndTime:      now,
			HasReceivedBids: false,
		},
	}

	// Prepare v015genstate
	s.v15genstate = v015auction.GenesisState{
		NextAuctionID: 10,
		Params:        v15params,
		Auctions: v015auction.GenesisAuctions{
			v015auction.SurplusAuction{
				BaseAuction: v015auction.BaseAuction{
					ID:              2,
					Initiator:       s.addresses[0].String(),
					Lot:             sdk.NewInt64Coin("kava", 12),
					Bidder:          s.addresses[1],
					Bid:             sdk.NewInt64Coin("kava", 12),
					EndTime:         now,
					MaxEndTime:      now,
					HasReceivedBids: false,
				},
			},
		},
	}

	expectedGenState, err := v016auction.NewGenesisState(10, v16params, []v016auction.GenesisAuction{expectedAuction})
	s.Require().NoError(err)

	genState := Migrate(s.v15genstate)
	s.Equal(expectedGenState, genState)
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
