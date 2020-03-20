package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
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
		msgTest{
			from:       sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			denom:      "bnb",
			expectPass: true,
		},
		msgTest{
			from:       sdk.AccAddress(crypto.AddressHash([]byte("KavaTest1"))),
			denom:      "",
			expectPass: false,
		},
		msgTest{
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
			suite.NoError(err)
		} else {
			suite.Error(err)
		}
	}
}

func TestMsgTestSuite(t *testing.T) {
	suite.Run(t, new(MsgTestSuite))
}
