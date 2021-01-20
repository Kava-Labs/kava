package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/hard/types"
)

type GenesisTestSuite struct {
	suite.Suite
}

func (suite *GenesisTestSuite) TestGenesisValidation() {
	type args struct {
		params types.Params
		pbt    time.Time
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
				pbt:    types.DefaultPreviousBlockTime,
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid",
			args: args{
				params: types.NewParams(true, types.DefaultMoneyMarkets, types.DefaultCheckLtvIndexCount),
				pbt:    time.Date(2020, 10, 8, 12, 0, 0, 0, time.UTC),
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "invalid previous blocktime",
			args: args{
				params: types.NewParams(true, types.DefaultMoneyMarkets, types.DefaultCheckLtvIndexCount),
				pbt:    time.Time{},
			},
			expectPass:  false,
			expectedErr: "previous block time not set",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			gs := types.NewGenesisState(tc.args.params, tc.args.pbt)
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
