package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var _ sdk.AnteDecorator = EvmMinGasFilter{}

// EVMKeeper specifies the interface that EvmMinGasFilter requires
type EVMKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
}

// EvmMinGasFilter filters out the EvmDenom min gas price and calls the next ante handle with an updated context
type EvmMinGasFilter struct {
	evmKeeper EVMKeeper
}

// NewEvmMinGasFilter takes an EVMKeeper and returns a new min gas filter for it's EvmDenom
func NewEvmMinGasFilter(evmKeeper EVMKeeper) EvmMinGasFilter {
	return EvmMinGasFilter{
		evmKeeper: evmKeeper,
	}
}

// AnteHandle checks the EvmDenom from the evmKeeper and filters out the EvmDenom from the ctx
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
