package accumulators

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

type IncentiveKeeper interface {
	SetRewardAccrualTime(
		ctx sdk.Context,
		claimType types.ClaimType,
		collateralType string,
		previousAccumulationTime time.Time,
	)
	GetRewardAccrualTime(
		ctx sdk.Context,
		claimType types.ClaimType,
		collateralType string,
	) (time.Time, bool)
	GetRewardIndexesOfClaimType(
		ctx sdk.Context,
		claimType types.ClaimType,
		collateralType string,
	) (types.RewardIndexes, bool)
	SetRewardIndexes(
		ctx sdk.Context,
		claimType types.ClaimType,
		collateralType string,
		indexes types.RewardIndexes,
	)
}
