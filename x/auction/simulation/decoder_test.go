package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/auction/types"
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
	auction := types.NewSurplusAuction("me", oneCoin, "coin", time.Now().UTC())

	kvPairs := kv.Pairs{
		kv.Pair{Key: types.AuctionKeyPrefix, Value: cdc.MustMarshalBinaryLengthPrefixed(&auction)},
		kv.Pair{Key: types.AuctionByTimeKeyPrefix, Value: sdk.Uint64ToBigEndian(2)},
		kv.Pair{Key: types.NextAuctionIDKey, Value: sdk.Uint64ToBigEndian(10)},
		kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Auction", fmt.Sprintf("%v\n%v", auction, auction)},
		{"AuctionByTime", "2\n2"},
		{"NextAuctionI", "10\n10"},
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
