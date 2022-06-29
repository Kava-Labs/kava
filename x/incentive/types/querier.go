package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the incentive module
const (
	QueryGetRewards            = "rewards"
	QueryGetHardRewards        = "hard-rewards"
	QueryGetUSDXMintingRewards = "usdx-minting-rewards"
	QueryGetDelegatorRewards   = "delegator-rewards"
	QueryGetSwapRewards        = "swap-rewards"
	QueryGetSavingsRewards     = "savings-rewards"
	QueryGetEarnRewards        = "earn-rewards"
	QueryGetRewardFactors      = "reward-factors"
	QueryGetParams             = "parameters"
	QueryGetAPYs               = "apys"

	RestClaimCollateralType = "collateral_type"
	RestClaimOwner          = "owner"
	RestClaimType           = "type"
	RestUnsynced            = "unsynced"
)

// QueryRewardsParams params for query /incentive/rewards/<claim type>
type QueryRewardsParams struct {
	Page           int            `json:"page" yaml:"page"`
	Limit          int            `json:"limit" yaml:"limit"`
	RewardType     RewardType     `json:"reward_type" yaml:"reward_type"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Unsynchronized bool           `json:"unsynchronized" yaml:"unsynchronized"`
}

// NewQueryRewardsParams returns QueryRewardsParams
func NewQueryRewardsParams(page, limit int, rewardType RewardType, owner sdk.AccAddress, unsynchronized bool) QueryRewardsParams {
	return QueryRewardsParams{
		Page:           page,
		Limit:          limit,
		RewardType:     rewardType,
		Owner:          owner,
		Unsynchronized: unsynchronized,
	}
}

// QueryGetRewardFactorsResponse holds the response to a reward factor query
type QueryGetRewardFactorsResponse struct {
	USDXMintingRewardFactors RewardIndexes      `json:"usdx_minting_reward_factors" yaml:"usdx_minting_reward_factors"`
	HardSupplyRewardFactors  MultiRewardIndexes `json:"hard_supply_reward_factors" yaml:"hard_supply_reward_factors"`
	HardBorrowRewardFactors  MultiRewardIndexes `json:"hard_borrow_reward_factors" yaml:"hard_borrow_reward_factors"`
	DelegatorRewardFactors   MultiRewardIndexes `json:"delegator_reward_factors" yaml:"delegator_reward_factors"`
	SwapRewardFactors        MultiRewardIndexes `json:"swap_reward_factors" yaml:"swap_reward_factors"`
	SavingsRewardFactors     MultiRewardIndexes `json:"savings_reward_factors" yaml:"savings_reward_factors"`
	EarnRewardFactors        MultiRewardIndexes `json:"earn_reward_factors" yaml:"earn_reward_factors"`
}

// NewQueryGetRewardFactorsResponse returns a new instance of QueryAllRewardFactorsResponse
func NewQueryGetRewardFactorsResponse(usdxMintingFactors RewardIndexes, supplyFactors,
	hardBorrowFactors, allFactors, savingsFactors, earnFactors MultiRewardIndexes,
) QueryGetRewardFactorsResponse {
	return QueryGetRewardFactorsResponse{
		USDXMintingRewardFactors: usdxMintingFactors,
		HardSupplyRewardFactors:  supplyFactors,
		HardBorrowRewardFactors:  hardBorrowFactors,
		SwapRewardFactors:        allFactors, // This now includes all normal factors
		SavingsRewardFactors:     savingsFactors,
		EarnRewardFactors:        earnFactors,
	}
}
