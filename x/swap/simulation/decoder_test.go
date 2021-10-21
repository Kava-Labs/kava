package simulation

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/kava-labs/kava/x/swap/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	return
}

func TestDecodeSwapStore(t *testing.T) {
	cdc := makeTestCodec()

	depositor := sdk.AccAddress(crypto.AddressHash([]byte("DepositorAddress")))
	reserves := sdk.NewCoins(
		sdk.NewCoin("ukava", sdk.NewInt(100000000)),
		sdk.NewCoin("usdx", sdk.NewInt(200000000)),
	)
	shares := sdk.NewInt(123456)

	poolRecord := types.NewPoolRecord(reserves, shares)
	shareRecord := types.NewShareRecord(depositor, poolRecord.PoolID, shares)

	kvPairs := kv.Pairs{
		kv.Pair{Key: types.PoolKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(poolRecord)},
		kv.Pair{Key: types.DepositorPoolSharesPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(shareRecord)},
		kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"PoolRecord", fmt.Sprintf("%v\n%v", poolRecord, poolRecord)},
		{"ShareRecord", fmt.Sprintf("%v\n%v", shareRecord, shareRecord)},
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
