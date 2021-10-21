package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/issuance/types"
)

type MsgTestSuite struct {
	suite.Suite

	addrs []sdk.AccAddress
}

func (suite *MsgTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, addrs := app.GeneratePrivKeyAddressPairs(2)
	suite.addrs = addrs
}

func (suite *MsgTestSuite) TestMsgIssueTokens() {
	type args struct {
		sender   sdk.AccAddress
		tokens   sdk.Coin
		receiver sdk.AccAddress
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"default",
			args{
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("valid", sdk.NewInt(100)),
				receiver: suite.addrs[1],
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid sender",
			args{
				sender:   sdk.AccAddress{},
				tokens:   sdk.NewCoin("valid", sdk.NewInt(100)),
				receiver: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "sender address cannot be empty",
			},
		},
		{
			"invalid receiver",
			args{
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("valid", sdk.NewInt(100)),
				receiver: sdk.AccAddress{},
			},
			errArgs{
				expectPass: false,
				contains:   "receiver address cannot be empty",
			},
		},
		{
			"invalid tokens",
			args{
				sender:   suite.addrs[0],
				tokens:   sdk.Coin{Denom: "Invalid", Amount: sdk.NewInt(100)},
				receiver: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "invalid tokens",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			testMsg := types.NewMsgIssueTokens(tc.args.sender, tc.args.tokens, tc.args.receiver)
			err := testMsg.ValidateBasic()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *MsgTestSuite) TestMsgRedeemTokens() {
	type args struct {
		sender sdk.AccAddress
		tokens sdk.Coin
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"default",
			args{
				sender: suite.addrs[0],
				tokens: sdk.NewCoin("valid", sdk.NewInt(100)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid sender",
			args{
				sender: sdk.AccAddress{},
				tokens: sdk.NewCoin("valid", sdk.NewInt(100)),
			},
			errArgs{
				expectPass: false,
				contains:   "sender address cannot be empty",
			},
		},
		{
			"invalid tokens",
			args{
				sender: suite.addrs[0],
				tokens: sdk.Coin{Denom: "Invalid", Amount: sdk.NewInt(100)},
			},
			errArgs{
				expectPass: false,
				contains:   "invalid tokens",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			testMsg := types.NewMsgRedeemTokens(tc.args.sender, tc.args.tokens)
			err := testMsg.ValidateBasic()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *MsgTestSuite) TestMsgBlockAddress() {
	type args struct {
		sender  sdk.AccAddress
		denom   string
		address sdk.AccAddress
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"default",
			args{
				sender:  suite.addrs[0],
				denom:   "valid",
				address: suite.addrs[1],
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid sender",
			args{
				sender:  sdk.AccAddress{},
				denom:   "valid",
				address: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "sender address cannot be empty",
			},
		},
		{
			"invalid blocked",
			args{
				sender:  suite.addrs[0],
				denom:   "valid",
				address: sdk.AccAddress{},
			},
			errArgs{
				expectPass: false,
				contains:   "blocked address cannot be empty",
			},
		},
		{
			"invalid denom",
			args{
				sender:  suite.addrs[0],
				denom:   "Invalid",
				address: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			testMsg := types.NewMsgBlockAddress(tc.args.sender, tc.args.denom, tc.args.address)
			err := testMsg.ValidateBasic()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *MsgTestSuite) TestMsgUnblockAddress() {
	type args struct {
		sender  sdk.AccAddress
		denom   string
		address sdk.AccAddress
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"default",
			args{
				sender:  suite.addrs[0],
				denom:   "valid",
				address: suite.addrs[1],
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid sender",
			args{
				sender:  sdk.AccAddress{},
				denom:   "valid",
				address: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "sender address cannot be empty",
			},
		},
		{
			"invalid blocked",
			args{
				sender:  suite.addrs[0],
				denom:   "valid",
				address: sdk.AccAddress{},
			},
			errArgs{
				expectPass: false,
				contains:   "blocked address cannot be empty",
			},
		},
		{
			"invalid denom",
			args{
				sender:  suite.addrs[0],
				denom:   "Invalid",
				address: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			testMsg := types.NewMsgUnblockAddress(tc.args.sender, tc.args.denom, tc.args.address)
			err := testMsg.ValidateBasic()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *MsgTestSuite) TestMsgSetPauseStatus() {
	type args struct {
		sender sdk.AccAddress
		denom  string
		status bool
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"default",
			args{
				sender: suite.addrs[0],
				denom:  "valid",
				status: true,
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid sender",
			args{
				sender: sdk.AccAddress{},
				denom:  "valid",
				status: true,
			},
			errArgs{
				expectPass: false,
				contains:   "sender address cannot be empty",
			},
		},
		{
			"invalid denom",
			args{
				sender: suite.addrs[0],
				denom:  "Invalid",
				status: true,
			},
			errArgs{
				expectPass: false,
				contains:   "invalid denom",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			testMsg := types.NewMsgSetPauseStatus(tc.args.sender, tc.args.denom, tc.args.status)
			err := testMsg.ValidateBasic()
			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}
