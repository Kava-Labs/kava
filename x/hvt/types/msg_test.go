package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hvt/types"
)

type MsgTestSuite struct {
	suite.Suite
}

func (suite *MsgTestSuite) TestMsgDeposit() {
	type args struct {
		depositor   sdk.AccAddress
		amount      sdk.Coin
		depositType string
	}
	addrs := []sdk.AccAddress{
		sdk.AccAddress("test1"),
		sdk.AccAddress("test2"),
	}
	testCases := []struct {
		name        string
		args        args
		expectPass  bool
		expectedErr string
	}{
		{
			name: "valid",
			args: args{
				depositor:   addrs[0],
				amount:      sdk.NewCoin("bnb", sdk.NewInt(10000000)),
				depositType: "lp",
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid2",
			args: args{
				depositor:   addrs[0],
				amount:      sdk.NewCoin("bnb", sdk.NewInt(10000000)),
				depositType: "LP",
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "invalid",
			args: args{
				depositor:   addrs[0],
				amount:      sdk.NewCoin("bnb", sdk.NewInt(10000000)),
				depositType: "cat",
			},
			expectPass:  false,
			expectedErr: "invalid deposit type",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.NewMsgDeposit(tc.args.depositor, tc.args.amount, tc.args.depositType)
			err := msg.ValidateBasic()
			if tc.expectPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.expectedErr))
			}
		})
	}
}

func (suite *MsgTestSuite) TestMsgClaim() {
	type args struct {
		sender      sdk.AccAddress
		denom       string
		depositType string
		multiplier  string
	}
	addrs := []sdk.AccAddress{
		sdk.AccAddress("test1"),
		sdk.AccAddress("test2"),
	}
	testCases := []struct {
		name        string
		args        args
		expectPass  bool
		expectedErr string
	}{
		{
			name: "valid",
			args: args{
				sender:      addrs[0],
				denom:       "bnb",
				depositType: "lp",
				multiplier:  "large",
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid2",
			args: args{
				sender:      addrs[0],
				denom:       "bnb",
				depositType: "stake",
				multiplier:  "small",
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid3",
			args: args{
				sender:      addrs[0],
				denom:       "bnb",
				depositType: "lp",
				multiplier:  "Medium",
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "invalid",
			args: args{
				sender:      addrs[0],
				denom:       "bnb",
				depositType: "lp",
				multiplier:  "huge",
			},
			expectPass:  false,
			expectedErr: "reward multiplier",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.NewMsgClaimReward(tc.args.sender, tc.args.denom, tc.args.depositType, tc.args.multiplier)
			err := msg.ValidateBasic()
			if tc.expectPass {
				suite.NoError(err)
			} else {
				suite.Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.expectedErr))
			}
		})
	}
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}
