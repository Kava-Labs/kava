package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/evmutil/types"
)

const (
	erc20BalanceOfMethod = "balanceOf"
)

// DeployTestERC20Contract deploys an ERC20 contract on the EVM as the
// module account and returns the address of the contract. This contract has
// minting permissions for the module account.
// Derived from tharsis/evmos
// https://github.com/tharsis/evmos/blob/ee54f496551df937915ff6f74a94732a35abc505/x/erc20/keeper/evm.go
func (k Keeper) DeployTestERC20Contract(
	ctx sdk.Context,
	name string,
	symbol string,
) (types.InternalEVMAddress, error) {
	ctorArgs, err := types.CustomERC20Contract.ABI.Pack(
		"", // Empty string for contract constructor
		name,
		symbol,
	)
	if err != nil {
		return types.InternalEVMAddress{}, sdkerrors.Wrapf(err, "token %v is invalid", name)
	}

	data := make([]byte, len(types.CustomERC20Contract.Bin)+len(ctorArgs))
	copy(
		data[:len(types.CustomERC20Contract.Bin)],
		types.CustomERC20Contract.Bin,
	)
	copy(
		data[len(types.CustomERC20Contract.Bin):],
		ctorArgs,
	)

	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleEVMAddress.Bytes())
	if err != nil {
		return types.InternalEVMAddress{}, err
	}

	contractAddr := crypto.CreateAddress(types.ModuleEVMAddress, nonce)
	_, err = k.CallEVMWithData(ctx, types.ModuleEVMAddress, nil, data, big.NewInt(0))
	if err != nil {
		return types.InternalEVMAddress{}, fmt.Errorf("failed to deploy ERC20 for %s: %w", name, err)
	}

	return types.NewInternalEVMAddress(contractAddr), nil
}

// MintERC20 mints the given amount of an ERC20 token to an address. This is
// unchecked and should only be called after permission and enabled ERC20 checks.
func (k Keeper) MintERC20(
	ctx sdk.Context,
	contractAddr types.InternalEVMAddress,
	receiver types.InternalEVMAddress,
	amount *big.Int,
) error {
	_, err := k.CallEVM(
		ctx,
		types.CustomERC20Contract.ABI,
		types.ModuleEVMAddress,
		contractAddr,
		"mint",
		// Mint ERC20 args
		receiver.Address,
		amount,
	)

	return err
}

func (k Keeper) QueryERC20BalanceOf(
	ctx sdk.Context,
	contractAddr types.InternalEVMAddress,
	account types.InternalEVMAddress,
) (*big.Int, error) {
	res, err := k.CallEVM(
		ctx,
		types.CustomERC20Contract.ABI,
		types.ModuleEVMAddress,
		contractAddr,
		erc20BalanceOfMethod,
		// balanceOf ERC20 args
		account.Address,
	)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		if res.VmError == vm.ErrExecutionReverted.Error() {
			// Unpacks revert
			return nil, evmtypes.NewExecErrorWithReason(res.Ret)
		}

		return nil, status.Error(codes.Internal, res.VmError)
	}

	anyOutput, err := types.CustomERC20Contract.ABI.Unpack(erc20BalanceOfMethod, res.Ret)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to unpack method %v response: %w",
			erc20BalanceOfMethod,
			err,
		)
	}

	if len(anyOutput) != 1 {
		return nil, fmt.Errorf(
			"invalid ERC20 %v call return outputs %v, expected %v",
			erc20BalanceOfMethod,
			len(anyOutput),
			1,
		)
	}

	bal, ok := anyOutput[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf(
			"invalid ERC20 return type %T, expected %T",
			anyOutput[0],
			&big.Int{},
		)
	}

	return bal, nil
}
