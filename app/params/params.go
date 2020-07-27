package params

// Simulation parameter constants
const (
	StakePerAccount           = "stake_per_account"
	InitiallyBondedValidators = "initially_bonded_validators"
)

// Default simulation operation weights for messages and gov proposals
const (
	DefaultWeightMsgPlaceBid              int = 20
	DefaultWeightMsgCreateAtomicSwap      int = 20
	DefaultWeightMsgUpdatePrices          int = 20
	DefaultWeightMsgCdp                   int = 20
	DefaultWeightMsgClaimReward           int = 20
	OpWeightSubmitCommitteeChangeProposal int = 20
)
