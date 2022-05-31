package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.AnteDecorator = ActivateAfterDecorator{}

// ActivateAfterDecorator wraps a ante decorator, disabling it before a block height.
// It can be used to modify the antehandler in asynchronous chain upgrades.
type ActivateAfterDecorator struct {
	WrappedDecorator sdk.AnteDecorator
	upgradeHeight    int64
}

func ActivateAfter(decorator sdk.AnteDecorator, upgradeHeight int64) ActivateAfterDecorator {
	return ActivateAfterDecorator{
		WrappedDecorator: decorator,
		upgradeHeight:    upgradeHeight,
	}
}

func (aad ActivateAfterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.BlockHeight() >= aad.upgradeHeight {
		return aad.WrappedDecorator.AnteHandle(ctx, tx, simulate, next)
	}
	return next(ctx, tx, simulate)
}
