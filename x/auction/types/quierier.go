package types

import (
	"strings"
)

const (
	// QueryGetAuction command for getting the information about a particular auction
	QueryGetAuction = "getauctions"
)

// QueryResAuctions Result Payload for an auctions query
type QueryResAuctions []string

// implement fmt.Stringer
func (n QueryResAuctions) String() string {
	return strings.Join(n[:], "\n")
}
