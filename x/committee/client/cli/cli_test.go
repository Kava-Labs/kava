package cli_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/client/cli"
)

type CLITestSuite struct {
	suite.Suite
	cdc *codec.Codec
}

func (suite *CLITestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.cdc = tApp.Codec()
}

func (suite *CLITestSuite) TestExampleCommitteeChangeProposal() {
	suite.NotPanics(func() { cli.MustGetExampleCommitteeChangeProposal(suite.cdc) })
}

func (suite *CLITestSuite) TestExampleCommitteeDeleteProposal() {
	suite.NotPanics(func() { cli.MustGetExampleCommitteeDeleteProposal(suite.cdc) })
}
func (suite *CLITestSuite) TestExampleParameterChangeProposal() {
	suite.NotPanics(func() { cli.MustGetExampleParameterChangeProposal(suite.cdc) })
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}
