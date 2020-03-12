package types

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

// GenerateSecureRandomNumber generates cryptographically strong pseudo-random number
func GenerateSecureRandomNumber() (*big.Int, error) {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(256), nil).Sub(max, big.NewInt(1)) // 256-bits integer i.e. 2^256 - 1

	// Generate number between 0 - max
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return big.NewInt(0), errors.New("random number generation error")
	}

	// Catch random numbers that encode to hexadecimal poorly
	if len(randomNumber.Text(16)) != 64 {
		return GenerateSecureRandomNumber()
	}

	return randomNumber, nil
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
	data = append(data, []byte(sender)...)
	data = append(data, []byte(senderOtherChain)...)
	return tmhash.Sum(data)
}
