package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
)

func TestKeeperDisableInflation(t *testing.T) {
	testFunc := func(ctx sdk.Context, k keeper.Keeper) {
		k.CheckAndDisableMintAndKavaDistInflation(ctx)
	}
	suite.Run(t, testutil.NewDisableInflationTestSuite(testFunc))
}
