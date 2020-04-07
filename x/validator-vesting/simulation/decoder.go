package simulation

import (
	"bytes"
	"fmt"
	"time"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/kava-labs/kava/x/validator-vesting/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding auth type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.ValidatorVestingAccountPrefix):
		var accA, accB exported.Account
		cdc.MustUnmarshalBinaryBare(kvA.Value, &accA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &accB)
		return fmt.Sprintf("%v\n%v", accA, accB)
	case bytes.Equal(kvA.Key, types.BlocktimeKey):
		var btA, btB time.Time
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &btA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &btB)
		return fmt.Sprintf("%v\n%v", btA, btB)
	default:
		panic(fmt.Sprintf("invalid %s key %X", types.ModuleName, kvA.Key))
	}
}
