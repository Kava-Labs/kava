package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewAnteHandler returns an 'AnteHandler' that will run actions before a tx is sent to a module's handler.
func NewAnteHandler(ak keeper.AccountKeeper, supplyKeeper types.SupplyKeeper, sigGasConsumer ante.SignatureVerificationGasConsumer, addressFetchers ...AddressFetcher) sdk.AnteHandler {
	decorators := []sdk.AnteDecorator{}

	decorators = append(decorators, ante.NewSetUpContextDecorator()) // outermost AnteDecorator. SetUpContext must be called first

	if len(addressFetchers) > 0 {
		decorators = append(decorators, NewAuthenticatedMempoolDecorator(addressFetchers...))
	}
	decorators = append(decorators,
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		ante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(ak),
		ante.NewDeductFeeDecorator(ak, supplyKeeper),
		ante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		ante.NewSigVerificationDecorator(ak),
		ante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
	)
	return sdk.ChainAnteDecorators(decorators...)
}
