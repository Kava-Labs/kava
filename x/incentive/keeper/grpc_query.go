package keeper

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/incentive/types"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
)

const (
	RewardTypeHard        = "hard"
	RewardTypeUSDXMinting = "usdx_minting"
	RewardTypeDelegator   = "delegator"
	RewardTypeSwap        = "swap"
	RewardTypeSavings     = "savings"
	RewardTypeEarn        = "earn"
)

type queryServer struct {
	keeper Keeper
}

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl creates a new server for handling gRPC queries.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{
		keeper: keeper,
	}
}

func (s queryServer) Params(
	ctx context.Context,
	req *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &types.QueryParamsResponse{
		Params: s.keeper.GetParams(sdkCtx),
	}, nil
}

func (s queryServer) Rewards(
	ctx context.Context,
	req *types.QueryRewardsRequest,
) (*types.QueryRewardsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	res := types.QueryRewardsResponse{}

	hasOwner := req.Owner != ""

	var owner sdk.AccAddress
	if hasOwner {
		addr, err := sdk.AccAddressFromBech32(req.Owner)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err)
		}

		owner = addr
	}

	if err := s.queryRewards(sdkCtx, &res, owner, hasOwner, req.RewardType); err != nil {
		return nil, err
	}

	if !req.Unsynchronized {
		if err := s.synchronizeRewards(sdkCtx, &res); err != nil {
			return nil, err
		}
	}

	return &res, nil
}

func (s queryServer) RewardFactors(
	ctx context.Context,
	req *types.QueryRewardFactorsRequest,
) (*types.QueryRewardFactorsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	var usdxFactors types.RewardIndexes
	s.keeper.IterateUSDXMintingRewardFactors(sdkCtx, func(collateralType string, factor sdk.Dec) (stop bool) {
		usdxFactors = usdxFactors.With(collateralType, factor)
		return false
	})

	var supplyFactors types.MultiRewardIndexes
	s.keeper.IterateHardSupplyRewardIndexes(sdkCtx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		supplyFactors = supplyFactors.With(denom, indexes)
		return false
	})

	var borrowFactors types.MultiRewardIndexes
	s.keeper.IterateHardBorrowRewardIndexes(sdkCtx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		borrowFactors = borrowFactors.With(denom, indexes)
		return false
	})

	var delegatorFactors types.MultiRewardIndexes
	s.keeper.IterateDelegatorRewardIndexes(sdkCtx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		delegatorFactors = delegatorFactors.With(denom, indexes)
		return false
	})

	var swapFactors types.MultiRewardIndexes
	s.keeper.IterateSwapRewardIndexes(sdkCtx, func(poolID string, indexes types.RewardIndexes) (stop bool) {
		swapFactors = swapFactors.With(poolID, indexes)
		return false
	})

	var savingsFactors types.MultiRewardIndexes
	s.keeper.IterateSavingsRewardIndexes(sdkCtx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		savingsFactors = savingsFactors.With(denom, indexes)
		return false
	})

	var earnFactors types.MultiRewardIndexes
	s.keeper.IterateEarnRewardIndexes(sdkCtx, func(denom string, indexes types.RewardIndexes) (stop bool) {
		earnFactors = earnFactors.With(denom, indexes)
		return false
	})

	return &types.QueryRewardFactorsResponse{
		UsdxMintingRewardFactors: usdxFactors,
		HardSupplyRewardFactors:  supplyFactors,
		HardBorrowRewardFactors:  borrowFactors,
		DelegatorRewardFactors:   delegatorFactors,
		SwapRewardFactors:        swapFactors,
		SavingsRewardFactors:     savingsFactors,
		EarnRewardFactors:        earnFactors,
	}, nil
}

func (s queryServer) Apy(
	ctx context.Context,
	req *types.QueryApyRequest,
) (*types.QueryApyResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params := s.keeper.GetParams(sdkCtx)
	var apys types.APYs

	// bkava APY (staking + incentive rewards)
	stakingAPR, err := GetStakingAPR(sdkCtx, s.keeper, params)
	if err != nil {
		return nil, err
	}

	apys = append(apys, types.NewAPY(liquidtypes.DefaultDerivativeDenom, stakingAPR))

	// Incentive only APYs
	for _, param := range params.EarnRewardPeriods {
		// Skip bkava as it's calculated earlier with staking rewards
		if param.CollateralType == liquidtypes.DefaultDerivativeDenom {
			continue
		}

		// Value in the vault in the same denom as CollateralType
		vaultTotalValue, err := s.keeper.earnKeeper.GetVaultTotalValue(sdkCtx, param.CollateralType)
		if err != nil {
			return nil, err
		}
		apy, err := GetAPYFromMultiRewardPeriod(sdkCtx, s.keeper, param.CollateralType, param, vaultTotalValue.Amount)
		if err != nil {
			return nil, err
		}

		apys = append(apys, types.NewAPY(param.CollateralType, apy))
	}

	return &types.QueryApyResponse{
		Earn: apys,
	}, nil
}

// queryRewards queries the rewards for a given owner and reward type, updating
// the response with the results in place.
func (s queryServer) queryRewards(
	ctx sdk.Context,
	res *types.QueryRewardsResponse,
	owner sdk.AccAddress,
	hasOwner bool,
	rewardType string,
) error {
	rewardType = strings.ToLower(rewardType)
	isAllRewards := rewardType == ""

	if !rewardTypeIsValid(rewardType) {
		return status.Errorf(codes.InvalidArgument, "invalid reward type for owner %s: %s", owner, rewardType)
	}

	if isAllRewards || rewardType == RewardTypeUSDXMinting {
		if hasOwner {
			usdxMintingClaim, foundUsdxMintingClaim := s.keeper.GetUSDXMintingClaim(ctx, owner)
			if foundUsdxMintingClaim {
				res.USDXMintingClaims = append(res.USDXMintingClaims, usdxMintingClaim)
			}
		} else {
			usdxMintingClaims := s.keeper.GetAllUSDXMintingClaims(ctx)
			res.USDXMintingClaims = append(res.USDXMintingClaims, usdxMintingClaims...)
		}
	}

	if isAllRewards || rewardType == RewardTypeHard {
		if hasOwner {
			hardClaim, foundHardClaim := s.keeper.GetHardLiquidityProviderClaim(ctx, owner)
			if foundHardClaim {
				res.HardLiquidityProviderClaims = append(res.HardLiquidityProviderClaims, hardClaim)
			}
		} else {
			hardClaims := s.keeper.GetAllHardLiquidityProviderClaims(ctx)
			res.HardLiquidityProviderClaims = append(res.HardLiquidityProviderClaims, hardClaims...)
		}
	}

	if isAllRewards || rewardType == RewardTypeDelegator {
		if hasOwner {
			delegatorClaim, foundDelegatorClaim := s.keeper.GetDelegatorClaim(ctx, owner)
			if foundDelegatorClaim {
				res.DelegatorClaims = append(res.DelegatorClaims, delegatorClaim)
			}
		} else {
			delegatorClaims := s.keeper.GetAllDelegatorClaims(ctx)
			res.DelegatorClaims = append(res.DelegatorClaims, delegatorClaims...)
		}
	}

	if isAllRewards || rewardType == RewardTypeSwap {
		if hasOwner {
			swapClaim, foundSwapClaim := s.keeper.GetSwapClaim(ctx, owner)
			if foundSwapClaim {
				res.SwapClaims = append(res.SwapClaims, swapClaim)
			}
		} else {
			swapClaims := s.keeper.GetAllSwapClaims(ctx)
			res.SwapClaims = append(res.SwapClaims, swapClaims...)
		}
	}

	if isAllRewards || rewardType == RewardTypeSavings {
		if hasOwner {
			savingsClaim, foundSavingsClaim := s.keeper.GetSavingsClaim(ctx, owner)
			if foundSavingsClaim {
				res.SavingsClaims = append(res.SavingsClaims, savingsClaim)
			}
		} else {
			savingsClaims := s.keeper.GetAllSavingsClaims(ctx)
			res.SavingsClaims = append(res.SavingsClaims, savingsClaims...)
		}
	}

	if isAllRewards || rewardType == RewardTypeEarn {
		if hasOwner {
			earnClaim, foundEarnClaim := s.keeper.GetEarnClaim(ctx, owner)
			if foundEarnClaim {
				res.EarnClaims = append(res.EarnClaims, earnClaim)
			}
		} else {
			earnClaims := s.keeper.GetAllEarnClaims(ctx)
			res.EarnClaims = append(res.EarnClaims, earnClaims...)
		}
	}

	return nil
}

// synchronizeRewards synchronizes all non-empty rewards in place.
func (s queryServer) synchronizeRewards(
	ctx sdk.Context,
	res *types.QueryRewardsResponse,
) error {
	// Synchronize all non-empty rewards
	for i, claim := range res.USDXMintingClaims {
		res.USDXMintingClaims[i] = s.keeper.SimulateUSDXMintingSynchronization(ctx, claim)
	}

	for i, claim := range res.HardLiquidityProviderClaims {
		res.HardLiquidityProviderClaims[i] = s.keeper.SimulateHardSynchronization(ctx, claim)
	}

	for i, claim := range res.DelegatorClaims {
		res.DelegatorClaims[i] = s.keeper.SimulateDelegatorSynchronization(ctx, claim)
	}

	for i, claim := range res.SwapClaims {
		syncedClaim, found := s.keeper.GetSynchronizedSwapClaim(ctx, claim.Owner)
		if !found {
			return status.Errorf(codes.Internal, "previously found swap claim for owner %s should still be found", claim.Owner)
		}
		res.SwapClaims[i] = syncedClaim
	}

	for i, claim := range res.SavingsClaims {
		syncedClaim, found := s.keeper.GetSynchronizedSavingsClaim(ctx, claim.Owner)
		if !found {
			return status.Errorf(codes.Internal, "previously found savings claim for owner %s should still be found", claim.Owner)
		}
		res.SavingsClaims[i] = syncedClaim
	}

	for i, claim := range res.EarnClaims {
		syncedClaim, found := s.keeper.GetSynchronizedEarnClaim(ctx, claim.Owner)
		if !found {
			return status.Errorf(codes.Internal, "previously found earn claim for owner %s should still be found", claim.Owner)
		}
		res.EarnClaims[i] = syncedClaim
	}

	return nil
}

func rewardTypeIsValid(rewardType string) bool {
	return rewardType == "" ||
		rewardType == RewardTypeHard ||
		rewardType == RewardTypeUSDXMinting ||
		rewardType == RewardTypeDelegator ||
		rewardType == RewardTypeSwap ||
		rewardType == RewardTypeSavings ||
		rewardType == RewardTypeEarn
}
