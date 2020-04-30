package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/crypto"

	"github.com/kava-labs/kava/x/incentive/types"
)

type msgTest struct {
	from       sdk.AccAddress
	denom      string
	expectPass bool
}

type MsgTestSuite struct {
	suite.Suite

	tests []msgTest
}

func (suite *MsgTestSuite) SetupTest() {
	tests := []msgTest{
		{
			from:       sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			denom:      "bnb",
			expectPass: true,
		},
		{
			from:       sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			denom:      "",
			expectPass: false,
		},
		{
			from:       sdk.AccAddress{},
			denom:      "bnb",
			expectPass: false,
		},
	}
	suite.tests = tests
}

func (suite *MsgTestSuite) TestMsgValidation() {
	for _, t := range suite.tests {
		msg := types.NewMsgClaimReward(t.from, t.denom)
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
