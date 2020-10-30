package types

// Event types for harvest module
const (
	EventTypeHarvestDeposit               = "harvest_deposit"
	EventTypeHarvestDelegatorDistribution = "harvest_delegator_distribution"
	EventTypeHarvestLPDistribution        = "harvest_lp_distribution"
	EventTypeDeleteHarvestDeposit         = "delete_harvest_deposit"
	EventTypeHarvestWithdrawal            = "harvest_withdrawal"
	EventTypeClaimHarvestReward           = "claim_harvest_reward"
	EventTypeHarvestBorrow                = "harvest_borrow"
	AttributeValueCategory                = ModuleName
	AttributeKeyBlockHeight               = "block_height"
	AttributeKeyRewardsDistribution       = "rewards_distributed"
	AttributeKeyDeposit                   = "deposit"
	AttributeKeyDepositType               = "deposit_type"
	AttributeKeyDepositDenom              = "deposit_denom"
	AttributeKeyDepositor                 = "depositor"
	AttributeKeyClaimHolder               = "claim_holder"
	AttributeKeyClaimAmount               = "claim_amount"
	AttributeKeyClaimMultiplier           = "claim_multiplier"
	AttributeKeyBorrow                    = "borrow"
	AttributeKeyBorrower                  = "borrower"
	AttributeKeyBorrowCoins               = "borrow_coins"
)
