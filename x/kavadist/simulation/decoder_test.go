package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/kava-labs/kava/x/kavadist/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"
)

func TestDecodeDistributionStore(t *testing.T) {
	prevBlockTime := time.Now().UTC()
	bPrevBlockTime, err := prevBlockTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{
				Key:   []byte(types.PreviousBlockTimeKey),
				Value: bPrevBlockTime,
			},
			{
				Key:   []byte{0x99},
				Value: []byte{0x99},
			},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"PreviousBlockTime", fmt.Sprintf("%s\n%s", prevBlockTime, prevBlockTime)},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { DecodeStore(kvPairs.GetPairs()[i], kvPairs.GetPairs()[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, DecodeStore(kvPairs.GetPairs()[i], kvPairs.GetPairs()[i]), tt.name)
			}
		})
	}
}
