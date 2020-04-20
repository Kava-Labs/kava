package params

// Simulation parameter constants
const (
	StakePerAccount           = "stake_per_account"
	InitiallyBondedValidators = "initially_bonded_validators"
)

// Default simulation operation weights for messages and gov proposals
const (
	DefaultWeightMsgPlaceBid         int = 100
	DefaultWeightMsgCreateAtomicSwap int = 100
	DefaultWeightMsgUpdatePrices     int = 100
	DefaultWeightMsgCdp              int = 100

	// DefaultWeightCommunitySpendProposal int = 5
	// DefaultWeightTextProposal           int = 5
	// DefaultWeightParamChangeProposal    int = 5
)
