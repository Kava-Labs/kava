package simulation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/bep3/types"
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	return
}

func TestDecodeDistributionStore(t *testing.T) {
	cdc := makeTestCodec()

	oneCoin := sdk.NewCoin("coin", sdk.OneInt())
	swap := types.NewAtomicSwap(sdk.Coins{oneCoin}, nil, 10, 100, nil, nil, "otherChainSender", "otherChainRec", 200, types.Completed, true, types.Outgoing)
	supply := types.AssetSupply{IncomingSupply: oneCoin, OutgoingSupply: oneCoin, CurrentSupply: oneCoin}
	bz := tmbytes.HexBytes([]byte{1, 2})

	kvPairs := kv.Pairs{
		kv.Pair{Key: types.AtomicSwapKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(swap)},
		kv.Pair{Key: types.AssetSupplyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(supply)},
		kv.Pair{Key: types.AtomicSwapByBlockPrefix, Value: bz},
		kv.Pair{Key: types.AtomicSwapByBlockPrefix, Value: bz},
		kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"AtomicSwap", fmt.Sprintf("%v\n%v", swap, swap)},
		{"AssetSupply", fmt.Sprintf("%v\n%v", supply, supply)},
		{"AtomicSwapByBlock", fmt.Sprintf("%s\n%s", bz, bz)},
		{"AtomicSwapLongtermStorage", fmt.Sprintf("%s\n%s", bz, bz)},
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
