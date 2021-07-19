package types_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
)

type ParamTestSuite struct {
	suite.Suite
}

func (suite *ParamTestSuite) SetupTest() {}

var rewardPeriodWithInvalidRewardsPerSecond = types.NewRewardPeriod(
	true,
	"bnb",
	time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
	time.Date(2024, 10, 15, 14, 0, 0, 0, time.UTC),
	sdk.Coin{Denom: "INVALID!@#😫", Amount: sdk.ZeroInt()},
)
var rewardMultiPeriodWithInvalidRewardsPerSecond = types.NewMultiRewardPeriod(
	true,
	"bnb",
	time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
	time.Date(2024, 10, 15, 14, 0, 0, 0, time.UTC),
	sdk.Coins{sdk.Coin{Denom: "INVALID!@#😫", Amount: sdk.ZeroInt()}},
)
var validMultiRewardPeriod = types.NewMultiRewardPeriod(
	true,
	"bnb",
	time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
	time.Date(2024, 10, 15, 14, 0, 0, 0, time.UTC),
	sdk.NewCoins(sdk.NewInt64Coin("swap", 1e9)),
)
var validRewardPeriod = types.NewRewardPeriod(
	true,
	"bnb-a",
	time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
	time.Date(2024, 10, 15, 14, 0, 0, 0, time.UTC),
	sdk.NewInt64Coin(types.USDXMintingRewardDenom, 1e9),
)

func (suite *ParamTestSuite) TestParamValidation() {
	type errArgs struct {
		expectPass bool
		contains   string
	}
	type test struct {
		name    string
		params  types.Params
		errArgs errArgs
	}

	testCases := []test{
		{
			"default is valid",
			types.DefaultParams(),
			errArgs{
				expectPass: true,
			},
		},
		{
			"valid",
			types.Params{
				USDXMintingRewardPeriods: types.RewardPeriods{
					types.NewRewardPeriod(
						true,
						"bnb-a",
						time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
						time.Date(2024, 10, 15, 14, 0, 0, 0, time.UTC),
						sdk.NewCoin(types.USDXMintingRewardDenom, sdk.NewInt(122354)),
					)},
				ClaimMultipliers: types.Multipliers{
					types.NewMultiplier(
						types.Small, 1, sdk.MustNewDecFromStr("0.25"),
					),
					types.NewMultiplier(
						types.Large, 1, sdk.MustNewDecFromStr("1.0"),
					),
				},
				HardSupplyRewardPeriods: types.DefaultMultiRewardPeriods,
				HardBorrowRewardPeriods: types.DefaultMultiRewardPeriods,
				DelegatorRewardPeriods:  types.DefaultMultiRewardPeriods,
				SwapRewardPeriods:       types.DefaultMultiRewardPeriods,
				ClaimEnd:                time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
			},
			errArgs{
				expectPass: true,
			},
		},
		{
			"invalid usdx minting period makes params invalid",
			types.Params{
				USDXMintingRewardPeriods: types.RewardPeriods{rewardPeriodWithInvalidRewardsPerSecond},
				HardSupplyRewardPeriods:  types.DefaultMultiRewardPeriods,
				HardBorrowRewardPeriods:  types.DefaultMultiRewardPeriods,
				DelegatorRewardPeriods:   types.DefaultMultiRewardPeriods,
				SwapRewardPeriods:        types.DefaultMultiRewardPeriods,
				ClaimMultipliers:         types.DefaultMultipliers,
				ClaimEnd:                 time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
			},
			errArgs{
				expectPass: false,
				contains:   fmt.Sprintf("reward denom must be %s", types.USDXMintingRewardDenom),
			},
		},
		{
			"invalid hard supply periods makes params invalid",
			types.Params{
				USDXMintingRewardPeriods: types.DefaultRewardPeriods,
				HardSupplyRewardPeriods:  types.MultiRewardPeriods{rewardMultiPeriodWithInvalidRewardsPerSecond},
				HardBorrowRewardPeriods:  types.DefaultMultiRewardPeriods,
				DelegatorRewardPeriods:   types.DefaultMultiRewardPeriods,
				SwapRewardPeriods:        types.DefaultMultiRewardPeriods,
				ClaimMultipliers:         types.DefaultMultipliers,
				ClaimEnd:                 time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
			},
			errArgs{
				expectPass: false,
				contains:   "invalid reward amount",
			},
		},
		{
			"invalid hard borrow periods makes params invalid",
			types.Params{
				USDXMintingRewardPeriods: types.DefaultRewardPeriods,
				HardSupplyRewardPeriods:  types.DefaultMultiRewardPeriods,
				HardBorrowRewardPeriods:  types.MultiRewardPeriods{rewardMultiPeriodWithInvalidRewardsPerSecond},
				DelegatorRewardPeriods:   types.DefaultMultiRewardPeriods,
				SwapRewardPeriods:        types.DefaultMultiRewardPeriods,
				ClaimMultipliers:         types.DefaultMultipliers,
				ClaimEnd:                 time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
			},
			errArgs{
				expectPass: false,
				contains:   "invalid reward amount",
			},
		},
		{
			"invalid delegator periods makes params invalid",
			types.Params{
				USDXMintingRewardPeriods: types.DefaultRewardPeriods,
				HardSupplyRewardPeriods:  types.DefaultMultiRewardPeriods,
				HardBorrowRewardPeriods:  types.DefaultMultiRewardPeriods,
				DelegatorRewardPeriods:   types.MultiRewardPeriods{rewardMultiPeriodWithInvalidRewardsPerSecond},
				SwapRewardPeriods:        types.DefaultMultiRewardPeriods,
				ClaimMultipliers:         types.DefaultMultipliers,
				ClaimEnd:                 time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
			},
			errArgs{
				expectPass: false,
				contains:   "invalid reward amount",
			},
		},
		{
			"invalid swap periods makes params invalid",
			types.Params{
				USDXMintingRewardPeriods: types.DefaultRewardPeriods,
				HardSupplyRewardPeriods:  types.DefaultMultiRewardPeriods,
				HardBorrowRewardPeriods:  types.DefaultMultiRewardPeriods,
				DelegatorRewardPeriods:   types.DefaultMultiRewardPeriods,
				SwapRewardPeriods:        types.MultiRewardPeriods{rewardMultiPeriodWithInvalidRewardsPerSecond},
				ClaimMultipliers:         types.DefaultMultipliers,
				ClaimEnd:                 time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
			},
			errArgs{
				expectPass: false,
				contains:   "invalid reward amount",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.params.Validate()

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func (suite *ParamTestSuite) TestRewardPeriods() {
	suite.Run("Validate", func() {
		type err struct {
			pass     bool
			contains string
		}
		testCases := []struct {
			name    string
			periods types.RewardPeriods
			expect  err
		}{
			{
				name: "single period is valid",
				periods: types.RewardPeriods{
					validRewardPeriod,
				},
				expect: err{
					pass: true,
				},
			},
			{
				name: "duplicated reward period is invalid",
				periods: types.RewardPeriods{
					validRewardPeriod,
					validRewardPeriod,
				},
				expect: err{
					contains: "duplicated reward period",
				},
			},
			{
				name: "invalid reward denom is invalid",
				periods: types.RewardPeriods{
					types.NewRewardPeriod(
						true,
						"bnb-a",
						time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
						time.Date(2024, 10, 15, 14, 0, 0, 0, time.UTC),
						sdk.NewInt64Coin("hard", 1e9),
					),
				},
				expect: err{
					contains: fmt.Sprintf("reward denom must be %s", types.USDXMintingRewardDenom),
				},
			},
		}
		for _, tc := range testCases {

			err := tc.periods.Validate()

			if tc.expect.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expect.contains)
			}
		}
	})
}

func (suite *ParamTestSuite) TestMultiRewardPeriods() {
	suite.Run("Validate", func() {
		type err struct {
			pass     bool
			contains string
		}
		testCases := []struct {
			name    string
			periods types.MultiRewardPeriods
			expect  err
		}{
			{
				name: "single period is valid",
				periods: types.MultiRewardPeriods{
					validMultiRewardPeriod,
				},
				expect: err{
					pass: true,
				},
			},
			{
				name: "duplicated reward period is invalid",
				periods: types.MultiRewardPeriods{
					validMultiRewardPeriod,
					validMultiRewardPeriod,
				},
				expect: err{
					contains: "duplicated reward period",
				},
			},
			{
				name: "invalid reward period is invalid",
				periods: types.MultiRewardPeriods{
					rewardMultiPeriodWithInvalidRewardsPerSecond,
				},
				expect: err{
					contains: "invalid reward amount",
				},
			},
		}
		for _, tc := range testCases {

			err := tc.periods.Validate()

			if tc.expect.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expect.contains)
			}
		}
	})
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}
