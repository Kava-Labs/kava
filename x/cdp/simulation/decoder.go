package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/cdp/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding cdp type
func DecodeStore(cdc *codec.Codec, kvA, kvB kv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.CdpIDKeyPrefix):
		var cdpIDsA, cdpIDsB []uint64
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &cdpIDsA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &cdpIDsB)
		return fmt.Sprintf("%v\n%v", cdpIDsA, cdpIDsB)

	case bytes.Equal(kvA.Key[:1], types.CdpKeyPrefix):
		var cdpA, cdpB types.CDP
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &cdpA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &cdpB)
		return fmt.Sprintf("%v\n%v", cdpA, cdpB)

	case bytes.Equal(kvA.Key[:1], types.CdpIDKey),
		bytes.Equal(kvA.Key[:1], types.CollateralRatioIndexPrefix):
		idA := binary.BigEndian.Uint64(kvA.Value)
		idB := binary.BigEndian.Uint64(kvB.Value)
		return fmt.Sprintf("%d\n%d", idA, idB)

	case bytes.Equal(kvA.Key[:1], types.DebtDenomKey),
		bytes.Equal(kvA.Key[:1], types.GovDenomKey):
		var denomA, denomB string
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &denomA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &denomB)
		return fmt.Sprintf("%s\n%s", denomA, denomB)

	case bytes.Equal(kvA.Key[:1], types.DepositKeyPrefix):
		var depositA, depositB types.Deposit
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &depositA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &depositB)
		return fmt.Sprintf("%s\n%s", depositA, depositB)

	case bytes.Equal(kvA.Key[:1], types.PrincipalKeyPrefix):
		var totalA, totalB sdk.Int
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &totalA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &totalB)
		return fmt.Sprintf("%s\n%s", totalA, totalB)

	case bytes.Equal(kvA.Key[:1], types.PreviousDistributionTimeKey):
		var timeA, timeB time.Time
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &timeA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &timeB)
		return fmt.Sprintf("%s\n%s", timeA, timeB)

	case bytes.Equal(kvA.Key[:1], types.InterestFactorPrefix),
		bytes.Equal(kvA.Key[:1], types.SavingsFactorPrefix):
		var decA, decB sdk.Dec
		cdc.MustUnmarshalBinaryBare(kvA.Value, &decA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &decB)
		return fmt.Sprintf("%s\n%s", decA, decB)
	case bytes.Equal(kvA.Key[:1], types.SavingsClaimsPrefix):
		var claimA, claimB types.USDXSavingsRateClaim
		cdc.MustUnmarshalBinaryBare(kvA.Value, &claimA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &claimB)
		return fmt.Sprintf("%s\n%s", claimA, claimB)
	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
