package simulation

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/kv"

	"github.com/kava-labs/kava/x/incentive/types"
)

// DecodeStore unmarshals the KVPair's Value to the module's corresponding type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {

	case bytes.Equal(kvA.Key[:1], types.USDXMintingClaimKeyPrefix):
		var claimA, claimB types.USDXMintingClaim
		cdc.MustUnmarshalBinaryBare(kvA.Value, &claimA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &claimB)
		return fmt.Sprintf("%v\n%v", claimA, claimB)

	case bytes.Equal(kvA.Key[:1], types.PreviousUSDXMintingRewardAccrualTimeKeyPrefix):
		var timeA, timeB time.Time
		cdc.MustUnmarshalBinaryBare(kvA.Value, &timeA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &timeB)
		return fmt.Sprintf("%s\n%s", timeA, timeB)

	case bytes.Equal(kvA.Key[:1], types.USDXMintingRewardFactorKeyPrefix):
		var factorA, factorB sdk.Dec
		cdc.MustUnmarshalBinaryBare(kvA.Value, &factorA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &factorB)
		return fmt.Sprintf("%s\n%s", factorA, factorB)

	// case bytes.Equal(kvA.Key[:1], types.HardLiquidityClaimKeyPrefix):
	// 	var claimA, claimB types.HardLiquidityProviderClaim
	// 	cdc.MustUnmarshalBinaryBare(kvA.Value, &claimA)
	// 	cdc.MustUnmarshalBinaryBare(kvB.Value, &claimB)
	// 	return fmt.Sprintf("%v\n%v", claimA, claimB)

	// case bytes.Equal(kvA.Key[:1], types.PreviousHardSupplyRewardAccrualTimeKeyPrefix):
	// 	var timeA, timeB time.Time
	// 	cdc.MustUnmarshalBinaryBare(kvA.Value, &timeA)
	// 	cdc.MustUnmarshalBinaryBare(kvB.Value, &timeB)
	// 	return fmt.Sprintf("%s\n%s", timeA, timeB)

	// case bytes.Equal(kvA.Key[:1], types.HardSupplyRewardFactorKeyPrefix):
	// 	var factorA, factorB sdk.Dec
	// 	cdc.MustUnmarshalBinaryBare(kvA.Value, &factorA)
	// 	cdc.MustUnmarshalBinaryBare(kvB.Value, &factorB)
	// 	return fmt.Sprintf("%s\n%s", factorA, factorB)

	// case bytes.Equal(kvA.Key[:1], types.PreviousHardBorrowRewardAccrualTimeKeyPrefix):
	// 	var timeA, timeB time.Time
	// 	cdc.MustUnmarshalBinaryBare(kvA.Value, &timeA)
	// 	cdc.MustUnmarshalBinaryBare(kvB.Value, &timeB)
	// 	return fmt.Sprintf("%s\n%s", timeA, timeB)

	// case bytes.Equal(kvA.Key[:1], types.HardSupplyRewardFactorKeyPrefix):
	// 	var factorA, factorB sdk.Dec
	// 	cdc.MustUnmarshalBinaryBare(kvA.Value, &factorA)
	// 	cdc.MustUnmarshalBinaryBare(kvB.Value, &factorB)
	// 	return fmt.Sprintf("%s\n%s", factorA, factorB)

	// case bytes.Equal(kvA.Key[:1], types.PreviousHardDelegatorRewardAccrualTimeKeyPrefix):
	// 	var timeA, timeB time.Time
	// 	cdc.MustUnmarshalBinaryBare(kvA.Value, &timeA)
	// 	cdc.MustUnmarshalBinaryBare(kvB.Value, &timeB)
	// 	return fmt.Sprintf("%s\n%s", timeA, timeB)

	// case bytes.Equal(kvA.Key[:1], types.HardDelegatorRewardFactorKeyPrefix):
	// 	var factorA, factorB sdk.Dec
	// 	cdc.MustUnmarshalBinaryBare(kvA.Value, &factorA)
	// 	cdc.MustUnmarshalBinaryBare(kvB.Value, &factorB)
	// 	return fmt.Sprintf("%s\n%s", factorA, factorB)

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
