package types

// Events for the module
const (
	AttributeValueCategory = ModuleName

	// Event Types
	EventTypeConvertERC20ToCoin = "convert_evm_erc20_to_coin"
	EventTypeConvertCoinToERC20 = "convert_evm_erc20_from_coin"

	// Event Attributes - Common
	AttributeKeyReceiver = "receiver"
	AttributeKeyAmount   = "amount"

	// Event Attributes - Conversions
	AttributeKeyInitiator    = "initiator"
	AttributeKeyERC20Address = "erc20_address"
)
