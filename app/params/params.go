package params

// Simulation parameter constants
const (
	StakePerAccount           = "stake_per_account"
	InitiallyBondedValidators = "initially_bonded_validators"
)

// Default simulation operation weights for messages and gov proposals
const (
	DefaultWeightMsgPlaceBid              int = 75
	DefaultWeightMsgCreateAtomicSwap      int = 50
	DefaultWeightMsgUpdatePrices          int = 50
	DefaultWeightMsgCdp                   int = 100
	DefaultWeightMsgClaimReward           int = 50
	OpWeightSubmitCommitteeChangeProposal int = 50
)
