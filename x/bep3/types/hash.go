package types

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/crypto/tmhash"
)

// GenerateSecureRandomNumber generates cryptographically strong pseudo-random number
func GenerateSecureRandomNumber() ([64]byte, error) {
	// Max is a 256-bits integer i.e. 2^256
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(256), nil)

	// Generate number in the range [0, max]
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return [64]byte{}, errors.New("random number generation error")
	}

	// Ensure length of 64 for hexadecimal encoding by padding with 0s
	var paddedNumber [64]byte
	copy(paddedNumber[:], fmt.Sprintf("%064x", randomNumber))
	return paddedNumber, nil
}

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
	data = append(data, sender.Bytes()...)
	data = append(data, []byte(senderOtherChain)...)
	return tmhash.Sum(data)
}
