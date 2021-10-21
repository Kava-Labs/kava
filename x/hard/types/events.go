package types

// Event types for hard module
const (
	EventTypeHardDeposit          = "hard_deposit"
	EventTypeHardWithdrawal       = "hard_withdrawal"
	EventTypeHardBorrow           = "hard_borrow"
	EventTypeHardLiquidation      = "hard_liquidation"
	EventTypeHardRepay            = "hard_repay"
	AttributeValueCategory        = ModuleName
	AttributeKeyDeposit           = "deposit"
	AttributeKeyDepositDenom      = "deposit_denom"
	AttributeKeyDepositCoins      = "deposit_coins"
	AttributeKeyDepositor         = "depositor"
	AttributeKeyBorrow            = "borrow"
	AttributeKeyBorrower          = "borrower"
	AttributeKeyBorrowCoins       = "borrow_coins"
	AttributeKeySender            = "sender"
	AttributeKeyRepayCoins        = "repay_coins"
	AttributeKeyLiquidatedOwner   = "liquidated_owner"
	AttributeKeyLiquidatedCoins   = "liquidated_coins"
	AttributeKeyKeeper            = "keeper"
	AttributeKeyKeeperRewardCoins = "keeper_reward_coins"
	AttributeKeyOwner             = "owner"
)
