package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	channelkeeper "github.com/cosmos/ibc-go/modules/core/04-channel/keeper"
	ibcante "github.com/cosmos/ibc-go/modules/core/ante"
)

// NewAnteHandler returns an 'AnteHandler' that will run actions before a tx is sent to a module's handler.
func NewAnteHandler(
	accountKeeper ante.AccountKeeper,
	bankKeeper types.BankKeeper,
	feegrantKeeper ante.FeegrantKeeper,
	ibcChannelKeeper channelkeeper.Keeper,
	signModeHandler authsigning.SignModeHandler, sigGasConsumer ante.SignatureVerificationGasConsumer, addressFetchers ...AddressFetcher) (sdk.AnteHandler, error) {
	if accountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if bankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if signModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	decorators := []sdk.AnteDecorator{}

	decorators = append(decorators,
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewRejectExtensionOptionsDecorator(),
	)
	if len(addressFetchers) > 0 {
		decorators = append(decorators, NewAuthenticatedMempoolDecorator(addressFetchers...))
	}
	decorators = append(decorators,
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(accountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(accountKeeper),
		ante.NewDeductFeeDecorator(accountKeeper, bankKeeper, feegrantKeeper),
		ante.NewSetPubKeyDecorator(accountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(accountKeeper),
		ante.NewSigGasConsumeDecorator(accountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(accountKeeper, signModeHandler),
		ante.NewIncrementSequenceDecorator(accountKeeper), // innermost AnteDecorator
		ibcante.NewAnteDecorator(ibcChannelKeeper),
	)
	return sdk.ChainAnteDecorators(decorators...), nil
}
