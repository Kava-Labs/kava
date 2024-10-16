package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hard/types"
)

const (
	USDX_CF = 1000000
	KAVA_CF = 1000000
	BTCB_CF = 100000000
	BNB_CF  = 100000000
	BUSD_CF = 100000000
)

type GenesisTestSuite struct {
	suite.Suite
}

func (suite *GenesisTestSuite) TestGenesisValidation() {
	type args struct {
		params types.Params
		gats   types.GenesisAccumulationTimes
		deps   types.Deposits
		brws   types.Borrows
		ts     sdk.Coins
		tb     sdk.Coins
		tr     sdk.Coins
	}
	testCases := []struct {
		name        string
		args        args
		expectPass  bool
		expectedErr string
	}{
		{
			name: "default",
			args: args{
				params: types.DefaultParams(),
				gats:   types.DefaultAccumulationTimes,
				deps:   types.DefaultDeposits,
				brws:   types.DefaultBorrows,
				ts:     types.DefaultTotalSupplied,
				tb:     types.DefaultTotalBorrowed,
				tr:     types.DefaultTotalReserves,
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid",
			args: args{
				params: types.NewParams(
					types.MoneyMarkets{
						types.NewMoneyMarket("usdx", types.NewBorrowLimit(true, sdkmath.LegacyMustNewDecFromStr("100000000000"), sdkmath.LegacyMustNewDecFromStr("1")), "usdx:usd", sdkmath.NewInt(USDX_CF), types.NewInterestRateModel(sdkmath.LegacyMustNewDecFromStr("0.05"), sdkmath.LegacyMustNewDecFromStr("2"), sdkmath.LegacyMustNewDecFromStr("0.8"), sdkmath.LegacyMustNewDecFromStr("10")), sdkmath.LegacyMustNewDecFromStr("0.05"), sdkmath.LegacyZeroDec()),
					},
					sdkmath.LegacyMustNewDecFromStr("10"),
				),
				gats: types.GenesisAccumulationTimes{
					types.NewGenesisAccumulationTime("usdx", time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC), sdkmath.LegacyOneDec(), sdkmath.LegacyOneDec()),
				},
				deps: types.DefaultDeposits,
				brws: types.DefaultBorrows,
				ts:   sdk.Coins{},
				tb:   sdk.Coins{},
				tr:   sdk.Coins{},
			},
			expectPass:  true,
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			gs := types.NewGenesisState(tc.args.params, tc.args.gats, tc.args.deps, tc.args.brws, tc.args.ts, tc.args.tb, tc.args.tr)
			err := gs.Validate()
			if tc.expectPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.expectedErr))
			}
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
