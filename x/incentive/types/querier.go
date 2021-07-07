package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Querier routes for the incentive module
const (
	QueryGetRewards                    = "rewards"
	QueryGetHardRewards                = "hard-rewards"
	QueryGetHardRewardsUnsynced        = "hard-rewards-unsynced"
	QueryGetUSDXMintingRewards         = "usdx-minting-rewards"
	QueryGetUSDXMintingRewardsUnsynced = "usdx-minting-rewards-unsynced"
	QueryGetDelegatorRewards           = "delegator-rewards"
	QueryGetDelegatorRewardsUnsynced   = "delegator-rewards-unsynced"
	QueryGetRewardFactors              = "reward-factors"
	QueryGetParams                     = "parameters"
	QueryGetRewardPeriods              = "reward-periods"
	QueryGetClaimPeriods               = "claim-periods"
	RestClaimCollateralType            = "collateral_type"
	RestClaimOwner                     = "owner"
	RestClaimType                      = "type"
	RestUnsynced                       = "unsynced"
)

// QueryRewardsParams params for query /incentive/rewards
type QueryRewardsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
	Type  string
}

// NewQueryRewardsParams returns QueryRewardsParams
func NewQueryRewardsParams(page, limit int, owner sdk.AccAddress, rewardType string) QueryRewardsParams {
	return QueryRewardsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
		Type:  rewardType,
	}
}

// QueryUSDXMintingRewardsParams params for query /incentive/rewards type usdx-minting
type QueryUSDXMintingRewardsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryUSDXMintingRewardsParams returns QueryUSDXMintingRewardsParams
func NewQueryUSDXMintingRewardsParams(page, limit int, owner sdk.AccAddress) QueryUSDXMintingRewardsParams {
	return QueryUSDXMintingRewardsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
	}
}

// QueryUSDXMintingRewardsUnsyncedParams params for query unsynced /incentive/rewards type usdx-minting
type QueryUSDXMintingRewardsUnsyncedParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryUSDXMintingRewardsUnsyncedParams returns QueryUSDXMintingRewardsUnsyncedParams
func NewQueryUSDXMintingRewardsUnsyncedParams(page, limit int, owner sdk.AccAddress) QueryUSDXMintingRewardsUnsyncedParams {
	return QueryUSDXMintingRewardsUnsyncedParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
	}
}

// QueryHardRewardsParams params for query /incentive/rewards type hard
type QueryHardRewardsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryHardRewardsParams returns QueryHardRewardsParams
func NewQueryHardRewardsParams(page, limit int, owner sdk.AccAddress) QueryHardRewardsParams {
	return QueryHardRewardsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
	}
}

// QueryHardRewardsUnsyncedParams params for query unsynced /incentive/rewards type hard
type QueryHardRewardsUnsyncedParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryHardRewardsUnsyncedParams returns QueryHardRewardsUnsyncedParams
func NewQueryHardRewardsUnsyncedParams(page, limit int, owner sdk.AccAddress) QueryHardRewardsUnsyncedParams {
	return QueryHardRewardsUnsyncedParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
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

// QueryDelegatorRewardsParams params for query /incentive/rewards type delegator
type QueryDelegatorRewardsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryDelegatorRewardsParams returns QueryDelegatorRewardsParams
func NewQueryDelegatorRewardsParams(page, limit int, owner sdk.AccAddress) QueryDelegatorRewardsParams {
	return QueryDelegatorRewardsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
	}
}

// QueryDelegatorRewardsUnsyncedParams params for query unsynced /incentive/rewards type delegator
type QueryDelegatorRewardsUnsyncedParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress
}

// NewQueryDelegatorRewardsUnsyncedParams returns QueryDelegatorRewardsUnsyncedParams
func NewQueryDelegatorRewardsUnsyncedParams(page, limit int, owner sdk.AccAddress) QueryDelegatorRewardsUnsyncedParams {
	return QueryDelegatorRewardsUnsyncedParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
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
