package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ sigVerifiableTx = (*authtypes.StdTx)(nil) // assert StdTx implements SigVerifiableTx

// SigVerifiableTx defines a Tx interface for all signature verification decorators
type sigVerifiableTx interface {
	GetSigners() []sdk.AccAddress
}

// AddressFetcher is a type signature for functions used by the AuthenticatedMempoolDecorator to get authorized addresses.
type AddressFetcher func(sdk.Context) []sdk.AccAddress

// AuthenticatedMempoolDecorator blocks all txs from reaching the mempool unless they're signed by on of the authorzed addresses.
// It only runs before entry to mempool (CheckTx), and not in consensus (DeliverTx)
type AuthenticatedMempoolDecorator struct {
	addressFetchers []AddressFetcher
}

func NewAuthenticatedMempoolDecorator(fetchers ...AddressFetcher) AuthenticatedMempoolDecorator {
	return AuthenticatedMempoolDecorator{
		addressFetchers: fetchers,
	}
}

func (amd AuthenticatedMempoolDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// This is only for local mempool purposes, and thus is only ran on check tx.
	if ctx.IsCheckTx() && !simulate {
		sigTx, ok := tx.(sigVerifiableTx)
		if !ok {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx must be sig verifiable tx")
		}
		if commonAddressesExist(sigTx.GetSigners(), amd.fetchAuthorizedAddresses(ctx)) {
			return ctx, fmt.Errorf("address not authorized for this mempool") // TODO
		}
	}
	return next(ctx, tx, simulate)
}

func (amd AuthenticatedMempoolDecorator) fetchAuthorizedAddresses(ctx sdk.Context) []sdk.AccAddress {
	addrs := []sdk.AccAddress{}
	for _, fetch := range amd.addressFetchers {
		addrs = append(addrs, fetch(ctx)...)
	}
	return addrs
}

// commonAddressesExist checks if there is any interection between two lists of addresses
func commonAddressesExist(addresses1, addresses2 []sdk.AccAddress) bool {
	for _, a1 := range addresses1 {
		for _, a2 := range addresses2 {
			if a1.Equals(a2) {
				return false
			}
		}
	}
	return true
}
