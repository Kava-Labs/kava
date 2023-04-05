package v0_16

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	app "github.com/kava-labs/kava/app"
	v015swap "github.com/kava-labs/kava/x/swap/legacy/v0_15"
	v016swap "github.com/kava-labs/kava/x/swap/types"
)

type migrateTestSuite struct {
	suite.Suite

	addresses   []sdk.AccAddress
	v15genstate v015swap.GenesisState
	cdc         codec.Codec
	legacyCdc   *codec.LegacyAmino
}

func (s *migrateTestSuite) SetupTest() {
	app.SetSDKConfig()

	s.v15genstate = v015swap.GenesisState{
		Params:       v015swap.Params{},
		PoolRecords:  v015swap.PoolRecords{},
		ShareRecords: v015swap.ShareRecords{},
	}

	config := app.MakeEncodingConfig()
	s.cdc = config.Marshaler

	legacyCodec := codec.NewLegacyAmino()
	s.legacyCdc = legacyCodec

	_, accAddresses := app.GeneratePrivKeyAddressPairs(10)
	s.addresses = accAddresses
}

func (s *migrateTestSuite) TestMigrate_JSON() {
	// Migrate v15 swap to v16
	file := filepath.Join("testdata", "v15-swap.json")
	data, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	err = s.legacyCdc.UnmarshalJSON(data, &s.v15genstate)
	s.Require().NoError(err)
	genstate := Migrate(s.v15genstate)

	// Compare expect v16 swap json with migrated json
	actual := s.cdc.MustMarshalJSON(genstate)
	file = filepath.Join("testdata", "v16-swap.json")
	expected, err := ioutil.ReadFile(file)
	s.Require().NoError(err)
	s.Require().JSONEq(string(expected), string(actual))
}

func (s *migrateTestSuite) TestMigrate_Params() {
	params := v015swap.Params{
		SwapFee: sdk.MustNewDecFromStr("0.33"),
		AllowedPools: v015swap.AllowedPools{
			{TokenA: "A", TokenB: "B"},
			{TokenA: "C", TokenB: "D"},
		},
	}
	expectedParams := v016swap.Params{
		SwapFee: sdk.MustNewDecFromStr("0.33"),
		AllowedPools: v016swap.AllowedPools{
			{TokenA: "A", TokenB: "B"},
			{TokenA: "C", TokenB: "D"},
		},
	}
	s.v15genstate.Params = params
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expectedParams, genState.Params)
}

func (s *migrateTestSuite) TestMigrate_PoolRecords() {
	s.v15genstate.PoolRecords = v015swap.PoolRecords{
		{
			PoolID:      "pool-1",
			ReservesA:   sdk.NewCoin("usdx", sdkmath.NewInt(100)),
			ReservesB:   sdk.NewCoin("xrpb", sdkmath.NewInt(200)),
			TotalShares: sdkmath.NewInt(300),
		},
		{
			PoolID:      "pool-2",
			ReservesA:   sdk.NewCoin("usdx", sdkmath.NewInt(500)),
			ReservesB:   sdk.NewCoin("ukava", sdkmath.NewInt(500)),
			TotalShares: sdkmath.NewInt(1000),
		},
	}
	expected := v016swap.PoolRecords{
		{
			PoolID:      "pool-1",
			ReservesA:   sdk.NewCoin("usdx", sdkmath.NewInt(100)),
			ReservesB:   sdk.NewCoin("xrpb", sdkmath.NewInt(200)),
			TotalShares: sdkmath.NewInt(300),
		},
		{
			PoolID:      "pool-2",
			ReservesA:   sdk.NewCoin("usdx", sdkmath.NewInt(500)),
			ReservesB:   sdk.NewCoin("ukava", sdkmath.NewInt(500)),
			TotalShares: sdkmath.NewInt(1000),
		},
	}
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expected, genState.PoolRecords)
}

func (s *migrateTestSuite) TestMigrate_ShareRecords() {
	s.v15genstate.ShareRecords = v015swap.ShareRecords{
		{
			PoolID:      "pool-1",
			Depositor:   s.addresses[0],
			SharesOwned: sdkmath.NewInt(100),
		},
		{
			PoolID:      "pool-2",
			Depositor:   s.addresses[1],
			SharesOwned: sdkmath.NewInt(410),
		},
	}
	expected := v016swap.ShareRecords{
		{
			PoolID:      "pool-1",
			Depositor:   s.addresses[0],
			SharesOwned: sdkmath.NewInt(100),
		},
		{
			PoolID:      "pool-2",
			Depositor:   s.addresses[1],
			SharesOwned: sdkmath.NewInt(410),
		},
	}
	genState := Migrate(s.v15genstate)
	s.Require().Equal(expected, genState.ShareRecords)
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(migrateTestSuite))
}
