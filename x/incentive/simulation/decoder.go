package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/kava-labs/kava/x/incentive/types"
)

// DecodeStore unmarshals the KVPair's Value to the module's corresponding type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.RewardPeriodKeyPrefix):
		var rewardPeriodA, rewardPeriodB types.RewardPeriod
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &rewardPeriodA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &rewardPeriodB)
		return fmt.Sprintf("%v\n%v", rewardPeriodA, rewardPeriodB)

	case bytes.Equal(kvA.Key[:1], types.ClaimPeriodKeyPrefix):
		var claimPeriodA, claimPeriodB types.ClaimPeriod
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &claimPeriodA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &claimPeriodB)
		return fmt.Sprintf("%v\n%v", claimPeriodA, claimPeriodB)

	case bytes.Equal(kvA.Key[:1], types.ClaimKeyPrefix):
		var claimA, claimB types.Claim
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &claimA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &claimB)
		return fmt.Sprintf("%v\n%v", claimA, claimB)

	case bytes.Equal(kvA.Key[:1], types.NextClaimPeriodIDPrefix):
		claimPeriodIDA := binary.BigEndian.Uint64(kvA.Value)
		claimPeriodIDB := binary.BigEndian.Uint64(kvB.Value)
		return fmt.Sprintf("%d\n%d", claimPeriodIDA, claimPeriodIDB)

	case bytes.Equal(kvA.Key[:1], types.PreviousBlockTimeKey):
		var timeA, timeB time.Time
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &timeA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &timeB)
		return fmt.Sprintf("%s\n%s", timeA, timeB)

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
