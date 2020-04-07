package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/kava-labs/kava/x/validator-vesting/types"
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

	acc := types.ValidatorVestingAccount{SigningThreshold: 1}
	now := time.Now().UTC()

	kvPairs := cmn.KVPairs{
		cmn.KVPair{Key: types.ValidatorVestingAccountPrefix, Value: cdc.MustMarshalBinaryBare(acc)},
		cmn.KVPair{Key: types.BlocktimeKey, Value: cdc.MustMarshalBinaryLengthPrefixed(now)},
		cmn.KVPair{Key: []byte{0x99}, Value: []byte{0x99}},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"ValidatorVestingAccount", fmt.Sprintf("%v\n%v", acc, acc)},
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
