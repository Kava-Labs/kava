package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the incentive module
const (
	QueryGetHardRewards        = "hard-rewards"
	QueryGetUSDXMintingRewards = "usdx-minting-rewards"
	QueryGetDelegatorRewards   = "delegator-rewards"
	QueryGetSwapRewards        = "swap-rewards"
	QueryGetRewardFactors      = "reward-factors"
	QueryGetParams             = "parameters"

	RestClaimCollateralType = "collateral_type"
	RestClaimOwner          = "owner"
	RestClaimType           = "type"
	RestUnsynced            = "unsynced"
)

// QueryRewardsParams params for query /incentive/rewards/<claim type>
type QueryRewardsParams struct {
	Page           int            `json:"page" yaml:"page"`
	Limit          int            `json:"limit" yaml:"limit"`
	Owner          sdk.AccAddress `json:"owner" yaml:"owner"`
	Unsynchronized bool           `json:"unsynchronized" yaml:"unsynchronized"`
}

// NewQueryRewardsParams returns QueryRewardsParams
func NewQueryRewardsParams(page, limit int, owner sdk.AccAddress, unsynchronized bool) QueryRewardsParams {
	return QueryRewardsParams{
		Page:           page,
		Limit:          limit,
		Owner:          owner,
		Unsynchronized: unsynchronized,
	}
}

// QueryRewardFactorsParams is the params for a filtered reward factors query
type QueryRewardFactorsParams struct {
	Denom string `json:"denom" yaml:"denom"`
}

// NewQueryRewardFactorsParams creates a new QueryRewardFactorsParams
func NewQueryRewardFactorsParams(denom string) QueryRewardFactorsParams {
	return QueryRewardFactorsParams{
		Denom: denom,
	}
}

// RewardFactor is a unique type returned by reward factor queries
type RewardFactor struct {
	Denom                   string        `json:"denom" yaml:"denom"`
	USDXMintingRewardFactor sdk.Dec       `json:"usdx_minting_reward_factor" yaml:"usdx_minting_reward_factor"`
	HardSupplyRewardFactors RewardIndexes `json:"hard_supply_reward_factors" yaml:"hard_supply_reward_factors"`
	HardBorrowRewardFactors RewardIndexes `json:"hard_borrow_reward_factors" yaml:"hard_borrow_reward_factors"`
	DelegatorRewardFactors  RewardIndexes `json:"delegator_reward_factors" yaml:"delegator_reward_factors"`
}

// NewRewardFactor returns a new instance of RewardFactor
func NewRewardFactor(denom string, usdxMintingRewardFactor sdk.Dec, hardSupplyRewardFactors,
	hardBorrowRewardFactors, delegatorRewardFactors RewardIndexes) RewardFactor {
	return RewardFactor{
		Denom:                   denom,
		USDXMintingRewardFactor: usdxMintingRewardFactor,
		HardSupplyRewardFactors: hardSupplyRewardFactors,
		HardBorrowRewardFactors: hardBorrowRewardFactors,
		DelegatorRewardFactors:  delegatorRewardFactors,
	}
}

// RewardFactors is a slice of RewardFactor
type RewardFactors = []RewardFactor
