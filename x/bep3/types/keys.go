package types

import (
	"encoding/hex"
)

const (
	// ModuleName is the name of the module
	ModuleName = "bep3"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute is the querier route for bep3
	QuerierRoute = ModuleName

	// DefaultParamspace default namestore
	DefaultParamspace = ModuleName
)

// Key prefixes
var (
	HTLTKeyPrefix = []byte{0x00} // prefix for keys that store HTLTs
)

// BytesToHexEncodedString converts data from []byte to a hex-encoded string
func BytesToHexEncodedString(data []byte) string {
	encodedData := make([]byte, hex.EncodedLen(len(data)))
	hex.Encode(encodedData, data)
	return string(encodedData)
}

// HexEncodedStringToBytes converts data from a hex-encoded string to []bytes
func HexEncodedStringToBytes(data string) ([]byte, error) {
	decodedData, err := hex.DecodeString(data)
	if err != nil {
		return []byte{}, err
	}
	return decodedData, nil
}
