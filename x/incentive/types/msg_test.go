package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/incentive/types"
)

type msgTest struct {
	from           sdk.AccAddress
	multiplierName string
	expectPass     bool
}

type MsgTestSuite struct {
	suite.Suite

	tests []msgTest
}

func (suite *MsgTestSuite) SetupTest() {
	tests := []msgTest{
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "large",
			expectPass:     true,
		},
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "medium",
			expectPass:     true,
		},
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "small",
			expectPass:     true,
		},
		{
			from:           sdk.AccAddress{},
			multiplierName: "medium",
			expectPass:     false,
		},
		{
			from:           sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			multiplierName: "huge",
			expectPass:     false,
		},
	}
	suite.tests = tests
}

func (suite *MsgTestSuite) TestMsgValidation() {
	for _, t := range suite.tests {
		msg := types.NewMsgClaimUSDXMintingReward(t.from, t.multiplierName)
		err := msg.ValidateBasic()
		if t.expectPass {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}
