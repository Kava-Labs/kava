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
		mms        types.MoneyMarkets
		ltvCounter int
		active     bool
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
				active: types.DefaultActive,
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid",
			args: args{
				mms:        types.DefaultMoneyMarkets,
				ltvCounter: 10,
				active:     true,
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "invalid rewards",
			args: args{
				mms:        types.DefaultMoneyMarkets,
				ltvCounter: 10,
				active:     true,
			},
			expectPass:  false,
			expectedErr: "reward denom should be hard",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := types.NewParams(tc.args.active, tc.args.mms, tc.args.ltvCounter)
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
