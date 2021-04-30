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
	QueryGetRewardFactors              = "reward-factors"
	QueryGetParams                     = "parameters"
	QueryGetRewardPeriods              = "reward-periods"
	QueryGetClaimPeriods               = "claim-periods"
	RestClaimCollateralType            = "collateral_type"
	RestClaimOwner                     = "owner"
	RestClaimDenom                     = "denom"
	RestClaimType                      = "type"
	RestUnsynced                       = "unsynced"
)

// QueryRewardsParams params for query /incentive/rewards
type QueryRewardsParams struct {
	Page  int            `json:"page" yaml:"page"`
	Limit int            `json:"limit" yaml:"limit"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	Denom string         `json:"denon" yaml:"denon"`
	Type  string         `json:"type" yaml:"type"`
}

// NewQueryRewardsParams returns QueryRewardsParams
func NewQueryRewardsParams(page, limit int, owner sdk.AccAddress, rewardType, denom string) QueryRewardsParams {
	return QueryRewardsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
		Type:  rewardType,
		Denom: denom,
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
	Denom string
}

// NewQueryHardRewardsParams returns QueryHardRewardsParams
func NewQueryHardRewardsParams(page, limit int, owner sdk.AccAddress, denom string) QueryHardRewardsParams {
	return QueryHardRewardsParams{
		Page:  page,
		Limit: limit,
		Owner: owner,
		Denom: denom,
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

// RewardFactor is a unique type returned by reward factor queries
type RewardFactor struct {
	Denom                     string        `json:"denom" yaml:"denom"`
	USDXMintingRewardFactor   sdk.Dec       `json:"usdx_minting_reward_factor" yaml:"usdx_minting_reward_factor"`
	HardSupplyRewardFactors   RewardIndexes `json:"hard_supply_reward_factors" yaml:"hard_supply_reward_factors"`
	HardBorrowRewardFactors   RewardIndexes `json:"hard_borrow_reward_factors" yaml:"hard_borrow_reward_factors"`
	HardDelegatorRewardFactor sdk.Dec       `json:"hard_delegator_reward_factor" yaml:"hard_delegator_reward_factor"`
}

// NewRewardFactor returns a new instance of RewardFactor
func NewRewardFactor(denom string, usdxMintingRewardFactor sdk.Dec, hardSupplyRewardFactors,
	hardBorrowRewardFactors RewardIndexes, hardDelegatorRewardFactor sdk.Dec) RewardFactor {
	return RewardFactor{
		Denom:                     denom,
		USDXMintingRewardFactor:   usdxMintingRewardFactor,
		HardSupplyRewardFactors:   hardSupplyRewardFactors,
		HardBorrowRewardFactors:   hardBorrowRewardFactors,
		HardDelegatorRewardFactor: hardDelegatorRewardFactor,
	}
}

// RewardFactors is a slice of RewardFactor
type RewardFactors = []RewardFactor
