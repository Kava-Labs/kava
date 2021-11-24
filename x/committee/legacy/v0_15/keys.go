package v0_15

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
