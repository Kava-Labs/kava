package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName The name that will be used throughout the module
	ModuleName = "committee"

	// StoreKey Top level store key where all module items will be stored
	StoreKey = ModuleName

	// RouterKey Top level router key
	RouterKey = ModuleName

	// QuerierRoute Top level query string
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

// Key prefixes
var (
	CommitteeKeyPrefix = []byte{0x00} // prefix for keys that store committees
	ProposalKeyPrefix  = []byte{0x01} // prefix for keys that store proposals
	VoteKeyPrefix      = []byte{0x02} // prefix for keys that store votes

	NextProposalIDKey = []byte{0x03} // key for the next proposal id
)

// GetKeyFromID returns the bytes to use as a key for a uint64 id
func GetKeyFromID(id uint64) []byte {
	return uint64ToBytes(id)
}

func GetVoteKey(proposalID uint64, voter sdk.AccAddress) []byte {
	return append(GetKeyFromID(proposalID), voter.Bytes()...)
}

// Uint64ToBytes converts a uint64 into fixed length bytes for use in store keys.
func uint64ToBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(id))
	return bz
}

// Uint64FromBytes converts some fixed length bytes back into a uint64.
func Uint64FromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
