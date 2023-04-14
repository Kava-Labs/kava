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

	// check all tx messages for a hard liquidate
	hasHardLiquidate := false
	for _, msg := range tx.GetMsgs() {
		if sdk.MsgTypeURL(msg) == sdk.MsgTypeURL(&hardtypes.MsgLiquidate{}) {
			hasHardLiquidate = true
			// one liquidate msg enough to end search
			break
		}
	}

	// continue if no hard liquidate messages are found
	if !hasHardLiquidate {
		return next(ctx, tx, simulate)
	}

	// reject transaction if greater than the set gas limit
	if feeTx.GetGas() > f.gasLimit {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%v", fmt.Errorf("liquidation gas %d over %d limit", feeTx.GetGas(), f.gasLimit))
	}

	// continue normally
	return next(ctx, tx, simulate)
}
