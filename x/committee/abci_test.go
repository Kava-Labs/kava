package committee_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
)

type ModuleTestSuite struct {
	suite.Suite

	keeper committee.Keeper
	app    app.TestApp
	ctx    sdk.Context

	addresses []sdk.AccAddress
}

func (suite *ModuleTestSuite) SetupTest() {
	suite.app = app.NewTestApp()
	suite.keeper = suite.app.GetCommitteeKeeper()
	suite.ctx = suite.app.NewContext(true, abci.Header{})
	_, suite.addresses = app.GeneratePrivKeyAddressPairs(5)
}

func (suite *ModuleTestSuite) TestBeginBlock() {
	suite.app.InitializeFromGenesisStates()
	// TODO replace below with genesis state
	normalCom := committee.Committee{
		ID:          12,
		Members:     suite.addresses[:2],
		Permissions: []committee.Permission{committee.GodPermission{}},
	}
	suite.keeper.SetCommittee(suite.ctx, normalCom)

	pprop1 := gov.NewTextProposal("1A Title", "A description of this proposal.")
	id1, err := suite.keeper.SubmitProposal(suite.ctx, normalCom.Members[0], normalCom.ID, pprop1)
	suite.NoError(err)

	oneHrLaterCtx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Hour))
	pprop2 := gov.NewTextProposal("2A Title", "A description of this proposal.")
	id2, err := suite.keeper.SubmitProposal(oneHrLaterCtx, normalCom.Members[0], normalCom.ID, pprop2)
	suite.NoError(err)

	// Run BeginBlocker
	proposalDurationLaterCtx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(committee.MaxProposalDuration))
	suite.NotPanics(func() {
		committee.BeginBlocker(proposalDurationLaterCtx, abci.RequestBeginBlock{}, suite.keeper)
	})

	// Check expired proposals are gone
	_, found := suite.keeper.GetProposal(suite.ctx, id1)
	suite.False(found, "expected expired proposal to be closed")
	_, found = suite.keeper.GetProposal(suite.ctx, id2)
	suite.True(found, "expected non expired proposal to be not closed")
}

func TestModuleTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}
