package ante

import (
	"fmt"

	hardtypes "github.com/kava-labs/kava/x/hard/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.AnteDecorator = HardLiquidateSpamFilter{}

// HardLiquidateSpamFIlter blocks hard liquidate messages over a gas limit from the mempool
type HardLiquidateSpamFilter struct {
	gasLimit uint64
}

// NewHardLiquidateSpamFIlter creates a hard liquidate spam filter to filter liquidate messages with inflated gas
func NewHardLiquidateSpamFilter(gasLimit uint64) HardLiquidateSpamFilter {
	return HardLiquidateSpamFilter{
		gasLimit: gasLimit,
	}
}

// AnteHandle implements sdk.AnteDecorator
func (f HardLiquidateSpamFilter) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// only run in checktx to filter mempool
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	// only apply to fee transactions
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return next(ctx, tx, simulate)
	}

	// check if the tx gas is over the limit
	isOverGasLimit := feeTx.GetGas() > f.gasLimit

	// tx is under limit, continue normally
	if !isOverGasLimit {
		return next(ctx, tx, simulate)
	}

	// tx is over limit and error if we find a hard liquidate message
	for _, msg := range tx.GetMsgs() {
		if sdk.MsgTypeURL(msg) == sdk.MsgTypeURL(&hardtypes.MsgLiquidate{}) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%v", fmt.Errorf("liquidation gas %d over %d limit", feeTx.GetGas(), f.gasLimit))
		}
	}

	// no hard liquidate messages found, continue normally
	return next(ctx, tx, simulate)
}
