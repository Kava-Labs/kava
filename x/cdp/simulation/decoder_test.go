package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tendermint/libs/common"

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
	deposit := types.Deposit{CdpID: 1, Amount: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))}
	principal := sdk.OneInt()
	prevDistTime := time.Now().UTC()

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: types.CdpIDKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(cdpIds)},
		cmn.KVPair{Key: types.CdpIDKey, Value: sdk.Uint64ToBigEndian(2)},
		cmn.KVPair{Key: types.CollateralRatioIndexPrefix, Value: sdk.Uint64ToBigEndian(10)},
		cmn.KVPair{Key: []byte(types.DebtDenomKey), Value: cdc.MustMarshalBinaryLengthPrefixed(denom)},
		cmn.KVPair{Key: []byte(types.GovDenomKey), Value: cdc.MustMarshalBinaryLengthPrefixed(denom)},
		cmn.KVPair{Key: []byte(types.DepositKeyPrefix), Value: cdc.MustMarshalBinaryLengthPrefixed(deposit)},
		cmn.KVPair{Key: []byte(types.PrincipalKeyPrefix), Value: cdc.MustMarshalBinaryLengthPrefixed(principal)},
		cmn.KVPair{Key: []byte(types.PreviousBlockTimeKey), Value: cdc.MustMarshalBinaryLengthPrefixed(prevDistTime)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"CdpIDs", fmt.Sprintf("%v\n%v", cdpIds, cdpIds)},
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
