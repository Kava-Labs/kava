package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

type paramTest struct {
	name      string
	params    types.Params
	errResult errResult
}

type errResult struct {
	expectPass bool
	contains   string
}

type ParamTestSuite struct {
	suite.Suite

	tests []paramTest
}

func (suite *ParamTestSuite) SetupTest() {
	suite.tests = []paramTest{
		{
			name: "valid - active",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			errResult: errResult{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "valid - inactive",
			params: types.Params{
				Active: false,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			errResult: errResult{
				expectPass: true,
				contains:   "",
			},
		},
		{
			name: "duplicate reward",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			errResult: errResult{
				expectPass: false,
				contains:   "cannot have duplicate reward collateral type",
			},
		},
		{
			name: "negative reward duration",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * -24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			errResult: errResult{
				expectPass: false,
				contains:   "reward duration must be positive",
			},
		},
		{
			name: "negative time lock",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, -1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			errResult: errResult{
				expectPass: false,
				contains:   "expected non-negative lockup",
			},
		},
		{
			name: "zero claim duration",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 0,
					},
				},
			},
			errResult: errResult{
				expectPass: false,
				contains:   "claim duration must be positive",
			},
		},
		{
			name: "zero reward",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "bnb-a",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(0)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			errResult: errResult{
				expectPass: false,
				contains:   "reward amount must be positive",
			},
		},
		{
			name: "empty reward collateral type",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						CollateralType:   "",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(1)),
						Duration:         time.Hour * 24 * 7,
						ClaimMultipliers: types.Multipliers{types.NewMultiplier(types.Small, 1, sdk.MustNewDecFromStr("0.33")), types.NewMultiplier(types.Large, 12, sdk.MustNewDecFromStr("1.0"))},
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			errResult: errResult{
				expectPass: false,
				contains:   "collateral type cannot be blank",
			},
		},
	}
}

func (suite *ParamTestSuite) TestParamValidation() {
	for _, t := range suite.tests {
		suite.Run(t.name, func() {
			err := t.params.Validate()
			if t.errResult.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), t.errResult.contains))
			}
		})
	}
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}
