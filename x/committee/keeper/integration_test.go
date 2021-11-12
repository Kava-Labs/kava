package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/testutil"
	"github.com/kava-labs/kava/x/committee/types"
)

// getProposalVoteMap collects up votes into a map indexed by proposalID
func getProposalVoteMap(k keeper.Keeper, ctx sdk.Context) map[uint64]([]types.Vote) {

	proposalVoteMap := map[uint64]([]types.Vote){}

	k.IterateProposals(ctx, func(p types.Proposal) bool {
		proposalVoteMap[p.ID] = k.GetVotesByProposal(ctx, p.ID)
		return false
	})
	return proposalVoteMap
}

func (suite *keeperTestSuite) getAccount(addr sdk.AccAddress) authtypes.AccountI {
	ak := suite.App.GetAccountKeeper()
	return ak.GetAccount(suite.Ctx, addr)
}

func mustNewTestMemberCommittee(addresses []sdk.AccAddress) *types.MemberCommittee {
	com, err := types.NewMemberCommittee(
		12,
		"This committee is for testing.",
		addresses,
		[]types.Permission{&types.GodPermission{}},
		testutil.D("0.667"),
		time.Hour*24*7,
		types.TALLY_OPTION_FIRST_PAST_THE_POST,
	)
	if err != nil {
		panic(err)
	}
	return com
}

// mustNewTestProposal returns a new test proposal.
func mustNewTestProposal() types.Proposal {
	proposal, err := types.NewProposal(
		govtypes.NewTextProposal("A Title", "A description of this proposal."),
		1, 1, time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		panic(err)
	}
	return proposal
}

// NewCommitteeGenesisState marshals a committee genesis state into json for use in initializing test apps.
func NewCommitteeGenesisState(cdc codec.Codec, gs *types.GenesisState) app.GenesisState {
	return app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(gs)}
}
