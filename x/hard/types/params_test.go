package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/hard/types"
)

type ParamTestSuite struct {
	suite.Suite
}

func (suite *ParamTestSuite) TestParamValidation() {
	type args struct {
		mms types.MoneyMarkets
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
				mms: types.DefaultMoneyMarkets,
			},
			expectPass:  true,
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := types.NewParams(tc.args.mms)
			err := params.Validate()
			if tc.expectPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.expectedErr))
			}
		})
	}
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}
