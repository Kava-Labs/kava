package types

const (
	EventTypeMintDerivative = "mint_derivative"
	EventTypeBurnDerivative = "burn_derivative"

	AttributeValueCategory        = ModuleName
	AttributeKeyDelegator         = "delegator"
	AttributeKeyValidator         = "validator"
	AttributeKeySharesTransferred = "shares_transferred"

	// TODO remove unused events
	AttributeKeyModuleAccount  = "module_account"
	AttributeKeyAmountReturned = "returned"
	AttributeKeyAmountBurned   = "burned"
)
