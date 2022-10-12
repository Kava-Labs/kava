package v2

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

type v1Params struct {
	USDXMintingRewardPeriods types.RewardPeriods        `protobuf:"bytes,1,rep,name=usdx_minting_reward_periods,json=usdxMintingRewardPeriods,proto3,castrepeated=RewardPeriods" json:"usdx_minting_reward_periods"`
	HardSupplyRewardPeriods  types.MultiRewardPeriods   `protobuf:"bytes,2,rep,name=hard_supply_reward_periods,json=hardSupplyRewardPeriods,proto3,castrepeated=MultiRewardPeriods" json:"hard_supply_reward_periods"`
	HardBorrowRewardPeriods  types.MultiRewardPeriods   `protobuf:"bytes,3,rep,name=hard_borrow_reward_periods,json=hardBorrowRewardPeriods,proto3,castrepeated=MultiRewardPeriods" json:"hard_borrow_reward_periods"`
	DelegatorRewardPeriods   types.MultiRewardPeriods   `protobuf:"bytes,4,rep,name=delegator_reward_periods,json=delegatorRewardPeriods,proto3,castrepeated=MultiRewardPeriods" json:"delegator_reward_periods"`
	SwapRewardPeriods        types.MultiRewardPeriods   `protobuf:"bytes,5,rep,name=swap_reward_periods,json=swapRewardPeriods,proto3,castrepeated=MultiRewardPeriods" json:"swap_reward_periods"`
	ClaimMultipliers         types.MultipliersPerDenoms `protobuf:"bytes,6,rep,name=claim_multipliers,json=claimMultipliers,proto3,castrepeated=MultipliersPerDenoms" json:"claim_multipliers"`
	ClaimEnd                 time.Time                  `protobuf:"bytes,7,opt,name=claim_end,json=claimEnd,proto3,stdtime" json:"claim_end"`
	SavingsRewardPeriods     types.MultiRewardPeriods   `protobuf:"bytes,8,rep,name=savings_reward_periods,json=savingsRewardPeriods,proto3,castrepeated=MultiRewardPeriods" json:"savings_reward_periods"`
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
func (p *v1Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(types.KeyUSDXMintingRewardPeriods, &p.USDXMintingRewardPeriods, validateRewardPeriodsParam),
		paramtypes.NewParamSetPair(types.KeyHardSupplyRewardPeriods, &p.HardSupplyRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(types.KeyHardBorrowRewardPeriods, &p.HardBorrowRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(types.KeyDelegatorRewardPeriods, &p.DelegatorRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(types.KeySwapRewardPeriods, &p.SwapRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(types.KeySavingsRewardPeriods, &p.SavingsRewardPeriods, validateMultiRewardPeriodsParam),
		paramtypes.NewParamSetPair(types.KeyMultipliers, &p.ClaimMultipliers, validateMultipliersPerDenomParam),
		paramtypes.NewParamSetPair(types.KeyClaimEnd, &p.ClaimEnd, validateClaimEndParam),
	}
}

// Validate checks that the parameters have valid values.
func (p v1Params) Validate() error {
	if err := validateMultipliersPerDenomParam(p.ClaimMultipliers); err != nil {
		return err
	}

	if err := validateRewardPeriodsParam(p.USDXMintingRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.HardSupplyRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.HardBorrowRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.DelegatorRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.SwapRewardPeriods); err != nil {
		return err
	}

	if err := validateMultiRewardPeriodsParam(p.SavingsRewardPeriods); err != nil {
		return err
	}

	return nil
}

func validateRewardPeriodsParam(i interface{}) error {
	rewards, ok := i.(types.RewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

func validateMultiRewardPeriodsParam(i interface{}) error {
	rewards, ok := i.(types.MultiRewardPeriods)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return rewards.Validate()
}

func validateMultipliersPerDenomParam(i interface{}) error {
	multipliers, ok := i.(types.MultipliersPerDenoms)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return multipliers.Validate()
}

func validateClaimEndParam(i interface{}) error {
	endTime, ok := i.(time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if endTime.Unix() <= 0 {
		return fmt.Errorf("end time should not be zero")
	}
	return nil
}
