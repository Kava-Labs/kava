package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
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

	currentPrice := types.CurrentPrice{MarketID: "current", Price: sdk.OneDec()}
	postedPrice := types.PostedPrice{MarketID: "posted", Price: sdk.OneDec(), Expiry: time.Now().UTC()}

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: []byte(types.CurrentPricePrefix), Value: cdc.MustMarshalBinaryBare(currentPrice)},
		cmn.KVPair{Key: []byte(types.RawPriceFeedPrefix), Value: cdc.MustMarshalBinaryBare(postedPrice)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"CurrentPrice", fmt.Sprintf("%v\n%v", currentPrice, currentPrice)},
		{"PostedPrice", fmt.Sprintf("%s\n%s", postedPrice, postedPrice)},
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
