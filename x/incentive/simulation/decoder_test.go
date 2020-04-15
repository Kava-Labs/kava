package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	return
}

func TestDecodeDistributionStore(t *testing.T) {
	cdc := makeTestCodec()

	// Set up RewardPeriod, ClaimPeriod, Claim, and previous block time
	rewardPeriod := types.NewRewardPeriod("btc", time.Now().UTC(), time.Now().Add(time.Hour*1).UTC(),
		sdk.NewCoin("ukava", sdk.NewInt(10000000000)), time.Now().Add(time.Hour*2).UTC(), time.Duration(time.Hour*2))
	claimPeriod := types.NewClaimPeriod("btc", 1, time.Now().Add(time.Hour*24).UTC(), time.Duration(time.Hour*24))
	addr, _ := sdk.AccAddressFromBech32("kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw")
	claim := types.NewClaim(addr, sdk.NewCoin("ukava", sdk.NewInt(1000000)), "bnb", 1)
	prevBlockTime := time.Now().Add(time.Hour * -1).UTC()

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: types.RewardPeriodKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&rewardPeriod)},
		cmn.KVPair{Key: types.ClaimPeriodKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&claimPeriod)},
		cmn.KVPair{Key: types.ClaimKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&claim)},
		cmn.KVPair{Key: types.NextClaimPeriodIDPrefix, Value: sdk.Uint64ToBigEndian(10)},
		cmn.KVPair{Key: []byte(types.PreviousBlockTimeKey), Value: cdc.MustMarshalBinaryLengthPrefixed(prevBlockTime)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"RewardPeriod", fmt.Sprintf("%v\n%v", rewardPeriod, rewardPeriod)},
		{"ClaimPeriod", fmt.Sprintf("%v\n%v", claimPeriod, claimPeriod)},
		{"Claim", fmt.Sprintf("%v\n%v", claim, claim)},
		{"NextClaimPeriodID", "10\n10"},
		{"PreviousBlockTime", fmt.Sprintf("%v\n%v", prevBlockTime, prevBlockTime)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStore(cdc, kvPairs[i], kvPairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStore(cdc, kvPairs[i], kvPairs[i]), tt.name)
			}
		})
	}
}
