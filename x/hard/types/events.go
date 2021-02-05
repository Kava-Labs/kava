package types

// Event types for hard module
const (
	EventTypeHardDeposit               = "hard_deposit"
	EventTypeHardDelegatorDistribution = "hard_delegator_distribution"
	EventTypeHardLPDistribution        = "hard_lp_distribution"
	EventTypeDeleteHardDeposit         = "delete_hard_deposit"
	EventTypeHardWithdrawal            = "hard_withdrawal"
	EventTypeHardBorrow                = "hard_borrow"
	EventTypeHardLiquidation           = "hard_liquidation"
	EventTypeHardRepay                 = "hard_repay"
	AttributeValueCategory             = ModuleName
	AttributeKeyBlockHeight            = "block_height"
	AttributeKeyRewardsDistribution    = "rewards_distributed"
	AttributeKeyDeposit                = "deposit"
	AttributeKeyDepositDenom           = "deposit_denom"
	AttributeKeyDepositCoins           = "deposit_coins"
	AttributeKeyDepositor              = "depositor"
	AttributeKeyBorrow                 = "borrow"
	AttributeKeyBorrower               = "borrower"
	AttributeKeyBorrowCoins            = "borrow_coins"
	AttributeKeySender                 = "sender"
	AttributeKeyRepayCoins             = "repay_coins"
	AttributeKeyLiquidatedOwner        = "liquidated_owner"
	AttributeKeyLiquidatedCoins        = "liquidated_coins"
	AttributeKeyKeeper                 = "keeper"
	AttributeKeyKeeperRewardCoins      = "keeper_reward_coins"
)
