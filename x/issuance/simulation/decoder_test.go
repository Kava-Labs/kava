package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/issuance/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	return
}

func TestDecodeIssuanceStore(t *testing.T) {
	cdc := makeTestCodec()
	supply := types.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour)
	prevBlockTime := time.Now().UTC()

	kvPairs := kv.Pairs{
		kv.Pair{Key: types.AssetSupplyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(supply)},
		kv.Pair{Key: []byte(types.PreviousBlockTimeKey), Value: cdc.MustMarshalBinaryLengthPrefixed(prevBlockTime)},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"AssetSupply", fmt.Sprintf("%v\n%v", supply, supply)},
		{"PreviousBlockTime", fmt.Sprintf("%s\n%s", prevBlockTime, prevBlockTime)},
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
