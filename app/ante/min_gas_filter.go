package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var _ sdk.AnteDecorator = EvmMinGasFilter{}

type EVMKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
}

type EvmMinGasFilter struct {
	evmKeeper EVMKeeper
}

func NewEvmMinGasFilter(evmKeeper EVMKeeper) EvmMinGasFilter {
	return EvmMinGasFilter{
		evmKeeper: evmKeeper,
	}
}

func (emgf EvmMinGasFilter) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	evmDenom := emgf.evmKeeper.GetParams(ctx).EvmDenom

	if ctx.MinGasPrices().AmountOf(evmDenom).IsPositive() {
		filteredMinGasPrices := sdk.NewDecCoins()

		for _, gasPrice := range ctx.MinGasPrices() {
			if gasPrice.Denom != evmDenom {
				filteredMinGasPrices = filteredMinGasPrices.Add(gasPrice)
			}
		}

		ctx = ctx.WithMinGasPrices(filteredMinGasPrices)
	}

	return next(ctx, tx, simulate)
}
