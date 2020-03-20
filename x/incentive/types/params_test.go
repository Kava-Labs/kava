package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
)

type paramTest struct {
	params     types.Params
	expectPass bool
}

type ParamTestSuite struct {
	suite.Suite

	tests []paramTest
}

func (suite *ParamTestSuite) SetupTest() {
	p1 := types.Params{
		Active: true,
		Rewards: types.Rewards{
			types.Reward{
				Active:        true,
				Denom:         "bnb",
				Reward:        sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
				Duration:      time.Hour * 24 * 7,
				TimeLock:      time.Hour * 8766,
				ClaimDuration: time.Hour * 24 * 14,
			},
		},
	}
	p2 := types.Params{
		Active: true,
		Rewards: types.Rewards{
			types.Reward{
				Active:        true,
				Denom:         "bnb",
				Reward:        sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
				Duration:      time.Hour * 24 * 7,
				TimeLock:      time.Hour * 8766,
				ClaimDuration: time.Hour * 24 * 14,
			},
			types.Reward{
				Active:        true,
				Denom:         "bnb",
				Reward:        sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
				Duration:      time.Hour * 24 * 7,
				TimeLock:      time.Hour * 8766,
				ClaimDuration: time.Hour * 24 * 14,
			},
		},
	}
	p3 := types.Params{
		Active: true,
		Rewards: types.Rewards{
			types.Reward{
				Active:        true,
				Denom:         "bnb",
				Reward:        sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
				Duration:      time.Hour * -24 * 7,
				TimeLock:      time.Hour * 8766,
				ClaimDuration: time.Hour * 24 * 14,
			},
		},
	}
	p4 := types.Params{
		Active: true,
		Rewards: types.Rewards{
			types.Reward{
				Active:        true,
				Denom:         "bnb",
				Reward:        sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
				Duration:      time.Hour * 24 * 7,
				TimeLock:      time.Hour * -8766,
				ClaimDuration: time.Hour * 24 * 14,
			},
		},
	}
	p5 := types.Params{
		Active: true,
		Rewards: types.Rewards{
			types.Reward{
				Active:        true,
				Denom:         "bnb",
				Reward:        sdk.NewCoin("ukava", sdk.NewInt(10000000000)),
				Duration:      time.Hour * 24 * 7,
				TimeLock:      time.Hour * 8766,
				ClaimDuration: time.Hour * 0,
			},
		},
	}
	p6 := types.Params{
		Active: true,
		Rewards: types.Rewards{
			types.Reward{
				Active:        true,
				Denom:         "bnb",
				Reward:        sdk.NewCoin("ukava", sdk.NewInt(0)),
				Duration:      time.Hour * 24 * 7,
				TimeLock:      time.Hour * 8766,
				ClaimDuration: time.Hour * 0,
			},
		},
	}

	suite.tests = []paramTest{
		paramTest{
			params:     p1,
			expectPass: true,
		},
		paramTest{
			params:     p2,
			expectPass: false,
		},
		paramTest{
			params:     p3,
			expectPass: false,
		},
		paramTest{
			params:     p4,
			expectPass: false,
		},
		paramTest{
			params:     p5,
			expectPass: false,
		},
		paramTest{
			params:     p6,
			expectPass: false,
		},
	}
}

func (suite *ParamTestSuite) TestParamValidation() {
	for _, t := range suite.tests {
		err := t.params.Validate()
		if t.expectPass {
			suite.NoError(err)
		} else {
			suite.Error(err)
		}
	}
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}
