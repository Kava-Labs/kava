package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/kava-labs/kava/x/auction/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding auction type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.AuctionKeyPrefix):
		var auctionA, auctionB types.Auction
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &auctionA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &auctionB)
		return fmt.Sprintf("%v\n%v", auctionA, auctionB)

	case bytes.Equal(kvA.Key[:1], types.AuctionByTimeKeyPrefix),
		bytes.Equal(kvA.Key[:1], types.NextAuctionIDKey):
		auctionIDA := binary.BigEndian.Uint64(kvA.Value)
		auctionIDB := binary.BigEndian.Uint64(kvB.Value)
		return fmt.Sprintf("%d\n%d", auctionIDA, auctionIDB)

	default:
		panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
	}
}
