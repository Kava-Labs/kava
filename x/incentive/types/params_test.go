package types_test

import (
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
				HardSupplyRewardPeriods: types.DefaultMultiRewardPeriods,
				HardBorrowRewardPeriods: types.DefaultMultiRewardPeriods,
				DelegatorRewardPeriods:  types.DefaultMultiRewardPeriods,
				SwapRewardPeriods:       types.DefaultMultiRewardPeriods,
				ClaimMultipliers: types.MultipliersPerDenom{
					{
						Denom: "hard",
						Multipliers: types.Multipliers{
							types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.25")),
							types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0")),
						},
					},
					{
						Denom: "ukava",
						Multipliers: types.Multipliers{
							types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.2")),
							types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0")),
						},
					},
				},
				ClaimEnd: time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
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
				contains:   "invalid reward amount",
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

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}
