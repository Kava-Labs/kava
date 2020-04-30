package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	types.RegisterCodec(cdc)
	return
}

func TestDecodeDistributionStore(t *testing.T) {
	cdc := makeTestCodec()

	cdpIds := []uint64{1, 2, 3, 4, 5}
	denom := "denom"
	oneCoins := sdk.NewCoin(denom, sdk.OneInt())
	deposit := types.Deposit{CdpID: 1, Amount: oneCoins}
	principal := sdk.OneInt()
	prevDistTime := time.Now().UTC()
	cdp := types.CDP{ID: 1, FeesUpdated: prevDistTime, Collateral: oneCoins, Principal: oneCoins, AccumulatedFees: oneCoins}

	kvPairs := kv.Pairs{
		kv.Pair{Key: types.CdpIDKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(cdpIds)},
		kv.Pair{Key: types.CdpKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(cdp)},
		kv.Pair{Key: types.CdpIDKey, Value: sdk.Uint64ToBigEndian(2)},
		kv.Pair{Key: types.CollateralRatioIndexPrefix, Value: sdk.Uint64ToBigEndian(10)},
		kv.Pair{Key: []byte(types.DebtDenomKey), Value: cdc.MustMarshalBinaryLengthPrefixed(denom)},
		kv.Pair{Key: []byte(types.GovDenomKey), Value: cdc.MustMarshalBinaryLengthPrefixed(denom)},
		kv.Pair{Key: []byte(types.DepositKeyPrefix), Value: cdc.MustMarshalBinaryLengthPrefixed(deposit)},
		kv.Pair{Key: []byte(types.PrincipalKeyPrefix), Value: cdc.MustMarshalBinaryLengthPrefixed(principal)},
		kv.Pair{Key: []byte(types.PreviousDistributionTimeKey), Value: cdc.MustMarshalBinaryLengthPrefixed(prevDistTime)},
		kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"CdpIDs", fmt.Sprintf("%v\n%v", cdpIds, cdpIds)},
		{"CDP", fmt.Sprintf("%v\n%v", cdp, cdp)},
		{"CdpID", "2\n2"},
		{"CollateralRatioIndex", "10\n10"},
		{"DebtDenom", fmt.Sprintf("%s\n%s", denom, denom)},
		{"GovDenom", fmt.Sprintf("%s\n%s", denom, denom)},
		{"DepositKeyPrefix", fmt.Sprintf("%v\n%v", deposit, deposit)},
		{"Principal", fmt.Sprintf("%v\n%v", principal, principal)},
		{"PreviousDistributionTime", fmt.Sprintf("%s\n%s", prevDistTime, prevDistTime)},
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
