package types

import (
	"encoding/binary"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// CalculateRandomHash calculates the hash of a number and timestamp
func CalculateRandomHash(randomNumber []byte, timestamp int64) []byte {
	data := make([]byte, RandomNumberLength+Int64Size)
	copy(data[:RandomNumberLength], randomNumber)
	binary.BigEndian.PutUint64(data[RandomNumberLength:], uint64(timestamp))
	return tmhash.Sum(data)
}

// CalculateSwapID calculates the hash of a RandomNumberHash, sdk.AccAddress, and string
func CalculateSwapID(randomNumberHash []byte, sender sdk.AccAddress, senderOtherChain string) []byte {
	senderOtherChain = strings.ToLower(senderOtherChain)
	data := randomNumberHash
	data = append(data, []byte(sender)...)
	data = append(data, []byte(senderOtherChain)...)
	return tmhash.Sum(data)
}
