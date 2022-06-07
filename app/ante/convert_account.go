package ante

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

type AccountKeeper interface {
	authante.AccountKeeper
}

var _ sdk.AnteDecorator = ConvertEthAccounts{}

// ConvertEthAccounts converts non contract eth accounts to base accounts, and calls the next ante handle with an updated context.
// This should run after signature verification to ensure only owners can convert accounts.
type ConvertEthAccounts struct {
	ak AccountKeeper
}

func NewConvertEthAccounts(ak AccountKeeper) ConvertEthAccounts {
	return ConvertEthAccounts{
		ak: ak,
	}
}

func (cea ConvertEthAccounts) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	signers := sigTx.GetSigners()
	for i := range signers {
		if err := convertAccount(ctx, cea.ak, signers[i]); err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func convertAccount(ctx sdk.Context, ak AccountKeeper, address sdk.AccAddress) error {
	acc, err := authante.GetSignerAcc(ctx, ak, address)
	if err != nil {
		return err
	}
	ethAcc, isEthAcc := acc.(*ethermint.EthAccount)
	if isEthAcc {
		isNotContract := bytes.Equal(ethAcc.GetCodeHash().Bytes(), evmtypes.EmptyCodeHash)
		if isNotContract {
			ak.SetAccount(ctx, ethAcc.BaseAccount)
		}
	}
	return nil
}
