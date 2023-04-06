package ante

import (
	"fmt"
	"runtime/debug"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibcante "github.com/cosmos/ibc-go/v6/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	evmante "github.com/evmos/ethermint/app/ante"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper, EVM Keeper and Fee Market Keeper.
type HandlerOptions struct {
	AccountKeeper          evmtypes.AccountKeeper
	BankKeeper             evmtypes.BankKeeper
	IBCKeeper              *ibckeeper.Keeper
	EvmKeeper              evmante.EVMKeeper
	FeegrantKeeper         authante.FeegrantKeeper
	SignModeHandler        authsigning.SignModeHandler
	SigGasConsumer         authante.SignatureVerificationGasConsumer
	FeeMarketKeeper        evmtypes.FeeMarketKeeper
	MaxTxGasWanted         uint64
	AddressFetchers        []AddressFetcher
	ExtensionOptionChecker authante.ExtensionOptionChecker
	TxFeeChecker           authante.TxFeeChecker
}

func (options HandlerOptions) Validate() error {
	if options.AccountKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.EvmKeeper == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "evm keeper is required for AnteHandler")
	}
	return nil
}

// cosmosHandlerOptions extends HandlerOptions to provide some Cosmos specific configurations
type cosmosHandlerOptions struct {
	HandlerOptions
	isEIP712 bool
}

// NewAnteHandler returns an 'AnteHandler' that will run actions before a tx is sent to a module's handler.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		defer Recover(ctx.Logger(), &err)

		txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 1 {
				return ctx, errorsmod.Wrap(
					sdkerrors.ErrInvalidRequest,
					"rejecting tx with more than 1 extension option",
				)
			}

			if len(opts) == 1 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case "/ethermint.evm.v1.ExtensionOptionsEthereumTx":
					// handle as *evmtypes.MsgEthereumTx
					anteHandler = newEthAnteHandler(options)
				case "/ethermint.types.v1.ExtensionOptionsWeb3Tx":
					// handle as normal Cosmos SDK tx, except signature is checked for EIP712 representation
					anteHandler = newCosmosAnteHandler(cosmosHandlerOptions{
						HandlerOptions: options,
						isEIP712:       true,
					})
				default:
					return ctx, errorsmod.Wrapf(
						sdkerrors.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s", typeURL,
					)
				}

				return anteHandler(ctx, tx, sim)
			}
		}

		// handle as totally normal Cosmos SDK tx
		switch tx.(type) {
		case sdk.Tx:
			anteHandler = newCosmosAnteHandler(cosmosHandlerOptions{
				HandlerOptions: options,
				isEIP712:       false,
			})
		default:
			return ctx, errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}, nil
}

func newCosmosAnteHandler(options cosmosHandlerOptions) sdk.AnteHandler {
	decorators := []sdk.AnteDecorator{}

	decorators = append(decorators,
		evmante.RejectMessagesDecorator{},   // reject MsgEthereumTxs
		authante.NewSetUpContextDecorator(), // second decorator. SetUpContext must be called before other decorators
	)

	if !options.isEIP712 {
		decorators = append(decorators, authante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker))
	}

	if len(options.AddressFetchers) > 0 {
		decorators = append(decorators, NewAuthenticatedMempoolDecorator(options.AddressFetchers...))
	}

	var sigVerification sdk.AnteDecorator = authante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler)
	if options.isEIP712 {
		sigVerification = evmante.NewLegacyEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler, options.EvmKeeper)
	}

	decorators = append(decorators,
		NewEvmMinGasFilter(options.EvmKeeper), // filter out evm denom from min-gas-prices
		NewVestingAccountDecorator(),
		NewAuthzLimiterDecorator(
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&vesting.MsgCreateVestingAccount{}),
			sdk.MsgTypeURL(&vesting.MsgCreatePermanentLockedAccount{}),
			sdk.MsgTypeURL(&vesting.MsgCreatePeriodicVestingAccount{}),
		),
		authante.NewValidateBasicDecorator(),
		authante.NewTxTimeoutHeightDecorator(),
		// If ethermint x/feemarket is enabled, align Cosmos min fee with the EVM
		// evmante.NewMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		authante.NewValidateMemoDecorator(options.AccountKeeper),
		authante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		authante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		authante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(options.AccountKeeper),
		authante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		sigVerification,
		authante.NewIncrementSequenceDecorator(options.AccountKeeper), // innermost AnteDecorator
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	)
	return sdk.ChainAnteDecorators(decorators...)
}

func newEthAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		evmante.NewEthSetUpContextDecorator(options.EvmKeeper), // outermost AnteDecorator. SetUpContext must be called first
		evmante.NewEthMempoolFeeDecorator(options.EvmKeeper),   // Check eth effective gas price against minimal-gas-prices
		evmante.NewEthValidateBasicDecorator(options.EvmKeeper),
		evmante.NewEthSigVerificationDecorator(options.EvmKeeper),
		evmante.NewEthAccountVerificationDecorator(options.AccountKeeper, options.EvmKeeper),
		evmante.NewCanTransferDecorator(options.EvmKeeper),
		evmante.NewEthGasConsumeDecorator(options.EvmKeeper, options.MaxTxGasWanted),
		evmante.NewEthIncrementSenderSequenceDecorator(options.AccountKeeper), // innermost AnteDecorator.
		evmante.NewEthEmitEventDecorator(options.EvmKeeper),                   // emit eth tx hash and index at the very last ante handler.
	)
}

func Recover(logger tmlog.Logger, err *error) {
	if r := recover(); r != nil {
		*err = errorsmod.Wrapf(sdkerrors.ErrPanic, "%v", r)

		if e, ok := r.(error); ok {
			logger.Error(
				"ante handler panicked",
				"error", e,
				"stack trace", string(debug.Stack()),
			)
		} else {
			logger.Error(
				"ante handler panicked",
				"recover", fmt.Sprintf("%v", r),
			)
		}
	}
}
