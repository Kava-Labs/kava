package types

// Events for the module
const (
	AttributeValueCategory = ModuleName

	// Event Types
	EventTypeConvertERC20ToCoin = "convert_erc20_to_coin"
	EventTypeConvertCoinToERC20 = "convert_coin_to_erc20"

	// Event Attributes - Common
	AttributeKeyReceiver = "receiver"
	AttributeKeyAmount   = "amount"

	// Event Attributes - Conversions
	AttributeKeyInitiator    = "initiator"
	AttributeKeyERC20Address = "erc20_address"
)
