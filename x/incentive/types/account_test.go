package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"

	"github.com/kava-labs/kava/x/incentive/types"
)

type accountTest struct {
	periods     vesting.Periods
	expectedVal int64
}

type AccountTestSuite struct {
	suite.Suite

	tests []accountTest
}

func (suite *AccountTestSuite) SetupTest() {
	tests := []accountTest{
		{
			periods: vesting.Periods{
				vesting.Period{
					Length: int64(100),
					Amount: sdk.Coins{},
				},
				vesting.Period{
					Length: int64(200),
					Amount: sdk.Coins{},
				},
			},
			expectedVal: int64(300),
		},
	}
	suite.tests = tests
}

func (suite *AccountTestSuite) TestGetTotalPeriodLength() {
	for _, t := range suite.tests {
		length := types.GetTotalVestingPeriodLength(t.periods)
		suite.Equal(t.expectedVal, length)
	}
}

func TestAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
