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
	EventTypeHarvestRepay                 = "harvest_repay"
	AttributeValueCategory                = ModuleName
	AttributeKeyBlockHeight               = "block_height"
	AttributeKeyRewardsDistribution       = "rewards_distributed"
	AttributeKeyDeposit                   = "deposit"
	AttributeKeyDepositDenom              = "deposit_denom"
	AttributeKeyDepositor                 = "depositor"
	AttributeKeyClaimType                 = "claim_type"
	AttributeKeyClaimHolder               = "claim_holder"
	AttributeKeyClaimAmount               = "claim_amount"
	AttributeKeyClaimMultiplier           = "claim_multiplier"
	AttributeKeyBorrow                    = "borrow"
	AttributeKeyBorrower                  = "borrower"
	AttributeKeyBorrowCoins               = "borrow_coins"
	AttributeKeySender                    = "sender"
	AttributeKeyRepayCoins                = "repay_coins"
)
