package v3

import (
	"fmt"

	"github.com/kava-labs/kava/x/incentive/types"
)

// Legacy store key prefixes
var (
	EarnClaimKeyPrefix                     = []byte{0x18} // prefix for keys that store earn claims
	EarnRewardIndexesKeyPrefix             = []byte{0x19} // prefix for key that stores earn reward indexes
	PreviousEarnRewardAccrualTimeKeyPrefix = []byte{0x20} // prefix for key that stores the previous time earn rewards accrued
)

func LegacyAccrualTimeKeyFromClaimType(claimType types.ClaimType) []byte {
	switch claimType {
	case types.CLAIM_TYPE_HARD_BORROW:
		panic("todo")
	case types.CLAIM_TYPE_HARD_SUPPLY:
		panic("todo")
	case types.CLAIM_TYPE_EARN:
		return PreviousEarnRewardAccrualTimeKeyPrefix
	case types.CLAIM_TYPE_SAVINGS:
		panic("todo")
	case types.CLAIM_TYPE_SWAP:
		panic("todo")
	case types.CLAIM_TYPE_USDX_MINTING:
		panic("todo")
	default:
		panic(fmt.Sprintf("unrecognized claim type: %s", claimType))
	}
}

func LegacyRewardIndexesKeyFromClaimType(claimType types.ClaimType) []byte {
	switch claimType {
	case types.CLAIM_TYPE_HARD_BORROW:
		panic("todo")
	case types.CLAIM_TYPE_HARD_SUPPLY:
		panic("todo")
	case types.CLAIM_TYPE_EARN:
		return EarnRewardIndexesKeyPrefix
	case types.CLAIM_TYPE_SAVINGS:
		panic("todo")
	case types.CLAIM_TYPE_SWAP:
		panic("todo")
	case types.CLAIM_TYPE_USDX_MINTING:
		panic("todo")
	default:
		panic(fmt.Sprintf("unrecognized claim type: %s", claimType))
	}
}
