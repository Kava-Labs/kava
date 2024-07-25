package params

// TODO this has not been updated for v0.44

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
	DefaultWeightMsgDeposit               int = 20
	DefaultWeightMsgWithdraw              int = 20
	DefaultWeightMsgSwapExactForTokens    int = 20
	DefaultWeightMsgSwapForExactTokens    int = 20
	DefaultWeightMsgIssue                 int = 20
	DefaultWeightMsgRedeem                int = 20
	DefaultWeightMsgBlock                 int = 20
	DefaultWeightMsgPause                 int = 20
	OpWeightSubmitCommitteeChangeProposal int = 20
)
