package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
)

type paramTest struct {
	name       string
	params     types.Params
	expectPass bool
}

type ParamTestSuite struct {
	suite.Suite

	tests []paramTest
}

func (suite *ParamTestSuite) SetupTest() {
	suite.tests = []paramTest{
		paramTest{
			name: "valid - active",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			expectPass: true,
		},
		paramTest{
			name: "valid - inactive",
			params: types.Params{
				Active: false,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			expectPass: true,
		},
		paramTest{
			name: "duplicate reward",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			expectPass: false,
		},
		paramTest{
			name: "negative reward duration",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * -24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			expectPass: false,
		},
		paramTest{
			name: "negative time lock",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * -8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			expectPass: false,
		},
		paramTest{
			name: "zero claim duration",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 0,
					},
				},
			},
			expectPass: false,
		},
		paramTest{
			name: "zero reward",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "bnb",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(0)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			expectPass: false,
		},
		paramTest{
			name: "empty reward denom",
			params: types.Params{
				Active: true,
				Rewards: types.Rewards{
					types.Reward{
						Active:           true,
						Denom:            "",
						AvailableRewards: sdk.NewCoin("ukava", sdk.NewInt(0)),
						Duration:         time.Hour * 24 * 7,
						TimeLock:         time.Hour * 8766,
						ClaimDuration:    time.Hour * 24 * 14,
					},
				},
			},
			expectPass: false,
		},
	}
}

func (suite *ParamTestSuite) TestParamValidation() {
	for _, t := range suite.tests {
		suite.Run(t.name, func() {
			err := t.params.Validate()
			if t.expectPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
			}
		})
	}
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}
