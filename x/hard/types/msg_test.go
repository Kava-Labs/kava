package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/hard/types"
)

type MsgTestSuite struct {
	suite.Suite
}

func (suite *MsgTestSuite) TestMsgDeposit() {
	type args struct {
		depositor sdk.AccAddress
		amount    sdk.Coins
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
				depositor: addrs[0],
				amount:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10000000))),
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid2",
			args: args{
				depositor: addrs[0],
				amount:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10000000))),
			},
			expectPass:  true,
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.NewMsgDeposit(tc.args.depositor, tc.args.amount)
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

func (suite *MsgTestSuite) TestMsgWithdraw() {
	type args struct {
		depositor sdk.AccAddress
		amount    sdk.Coins
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
				depositor: addrs[0],
				amount:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10000000))),
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "valid2",
			args: args{
				depositor: addrs[0],
				amount:    sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(10000000))),
			},
			expectPass:  true,
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.NewMsgWithdraw(tc.args.depositor, tc.args.amount)
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

func (suite *MsgTestSuite) TestMsgBorrow() {
	type args struct {
		borrower sdk.AccAddress
		amount   sdk.Coins
	}
	addrs := []sdk.AccAddress{
		sdk.AccAddress("test1"),
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
				borrower: addrs[0],
				amount:   sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1000000))),
			},
			expectPass:  true,
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.NewMsgBorrow(tc.args.borrower, tc.args.amount)
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

func (suite *MsgTestSuite) TestMsgRepay() {
	type args struct {
		sender sdk.AccAddress
		owner  sdk.AccAddress
		amount sdk.Coins
	}
	addrs := []sdk.AccAddress{
		sdk.AccAddress("test1"),
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
				sender: addrs[0],
				owner:  addrs[0],
				amount: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1000000))),
			},
			expectPass:  true,
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := types.NewMsgRepay(tc.args.sender, tc.args.owner, tc.args.amount)
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
