package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	evmtypes "github.com/tharsis/ethermint/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	emptyCodeHash = crypto.Keccak256(nil)
)

// SetEthAccountDecorator converts base accounts with pubkey type ethsecp256k1 to EthAccount when they transact on the cosmos-sdk
type SetEthAccountDecorator struct {
	ak authante.AccountKeeper
}

func NewSetEthAccountDecorator(ak authante.AccountKeeper) SetEthAccountDecorator {
	return SetEthAccountDecorator{
		ak: ak,
	}
}

// AnteHandle checks if each signer has a base account with pubkey type ethsecp256k1 and converts the account to EthAccount if so
func (sead SetEthAccountDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	pubkeys, err := sigTx.GetPubKeys()
	if err != nil {
		return ctx, err
	}
	signers := sigTx.GetSigners()

	for i := range pubkeys {
		acc, err := authante.GetSignerAcc(ctx, sead.ak, signers[i])
		if err != nil {
			return ctx, err
		}
		pk := acc.GetPubKey()
		_, isEthPubkey := pk.(*ethsecp256k1.PubKey)
		bacc, isBaseAcc := acc.(*authtypes.BaseAccount)
		if isEthPubkey && isBaseAcc {
			ethAcc := &evmtypes.EthAccount{
				BaseAccount: bacc,
				CodeHash:    common.BytesToHash(emptyCodeHash).String(),
			}
			sead.ak.SetAccount(ctx, ethAcc)
		}
	}
	return next(ctx, tx, simulate)
}

// SetEthAccountDecorator converts base accounts to EthAccount when they transact on the evm
type SetEthAccountDecoratorEVM struct {
	ak authante.AccountKeeper
}

func NewSetEthAccountDecoratorEVM(ak authante.AccountKeeper) SetEthAccountDecoratorEVM {
	return SetEthAccountDecoratorEVM{
		ak: ak,
	}
}

// AnteHandle checks if each signer has a base account and converts the account to EthAccount if so
func (sead SetEthAccountDecoratorEVM) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		// for all signers, convert BaseAccount to EthAccount
		for _, addr := range msg.GetSigners() {
			acc := sead.ak.GetAccount(ctx, addr)

			if acc == nil {
				return ctx, sdkerrors.Wrapf(
					sdkerrors.ErrUnknownAddress,
					"account %s (%s) is nil", common.BytesToAddress(addr.Bytes()), addr,
				)
			}
			bacc, isBaseAcc := acc.(*authtypes.BaseAccount)
			if isBaseAcc {
				ethAcc := &evmtypes.EthAccount{
					BaseAccount: bacc,
					CodeHash:    common.BytesToHash(emptyCodeHash).String(),
				}
				sead.ak.SetAccount(ctx, ethAcc)
			}
		}
	}
	return next(ctx, tx, simulate)
}
