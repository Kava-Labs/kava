package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// Querier routes for the incentive module
const (
	QueryGetRewards            = "rewards"
	QueryGetHardRewards        = "hard-rewards"
	QueryGetUSDXMintingRewards = "usdx-minting-rewards"
	QueryGetParams             = "parameters"
	QueryGetRewardPeriods      = "reward-periods"
	QueryGetClaimPeriods       = "claim-periods"
	RestClaimCollateralType    = "collateral_type"
	RestClaimOwner             = "owner"
	RestClaimType              = "type"
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

// PostClaimReq defines the properties of claim transaction's request body.
type PostClaimReq struct {
	BaseReq        rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Sender         sdk.AccAddress `json:"sender" yaml:"sender"`
	MultiplierName string         `json:"multiplier_name" yaml:"multiplier_name"`
}
