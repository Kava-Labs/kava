package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/committee"
	"github.com/kava-labs/kava/x/committee/keeper"
	"github.com/kava-labs/kava/x/committee/types"
)

// Avoid cluttering test cases with long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func d(str string) sdk.Dec                  { return sdk.MustNewDecFromStr(str) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }

// getProposalVoteMap collects up votes into a map indexed by proposalID
func getProposalVoteMap(k keeper.Keeper, ctx sdk.Context) map[uint64]([]types.Vote) {

	proposalVoteMap := map[uint64]([]types.Vote){}

	k.IterateProposals(ctx, func(p types.Proposal) bool {
		proposalVoteMap[p.ID] = k.GetVotesByProposal(ctx, p.ID)
		return false
	})
	return proposalVoteMap
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

// NewCommitteeGenesisState marshals a committee genesis state into json for use in initializing test apps.
func NewCommitteeGenesisState(cdc *codec.Codec, gs committee.GenesisState) app.GenesisState {
	return app.GenesisState{committee.ModuleName: cdc.MustMarshalJSON(gs)}
}
