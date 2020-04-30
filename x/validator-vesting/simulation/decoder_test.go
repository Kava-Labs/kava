package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/kv"

	"github.com/kava-labs/kava/x/validator-vesting/types"
)

var (
	pk1   = ed25519.GenPrivKey().PubKey()
	addr1 = sdk.AccAddress(pk1.Address())
)

func makeTestCodec() (cdc *codec.Codec) {
	cdc = codec.New()
	sdk.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	types.RegisterCodec(cdc)
	codec.RegisterEvidences(cdc)
	return
}

func TestDecodeDistributionStore(t *testing.T) {
	cdc := makeTestCodec()

	now := time.Now().UTC()

	kvPairs := kv.Pairs{
		kv.Pair{Key: append(types.ValidatorVestingAccountPrefix, addr1.Bytes()...), Value: []byte{0}},
		kv.Pair{Key: types.BlocktimeKey, Value: cdc.MustMarshalBinaryLengthPrefixed(now)},
		kv.Pair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ValidatorVestingAccount", fmt.Sprintf("%v\n%v", addr1, addr1)},
		{"BlockTime", fmt.Sprintf("%s\n%s", now, now)},
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
