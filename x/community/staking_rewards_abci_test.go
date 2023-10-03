package community_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/community"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/stretchr/testify/suite"
)

func TestABCIPayoutAccumulatedStakingRewards(t *testing.T) {
	testFunc := func(ctx sdk.Context, k keeper.Keeper) {
		community.BeginBlocker(ctx, k)
	}
	suite.Run(t, testutil.NewStakingRewardsTestSuite(testFunc))
}
