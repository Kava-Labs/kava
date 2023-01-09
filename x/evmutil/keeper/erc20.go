package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/evmutil/contract"
	"github.com/kava-labs/kava/x/evmutil/types"
)

const (
	erc20BalanceOfMethod = "balanceOf"
)

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
		contract.CustomERC20Contract.ABI,
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
		contract.CustomERC20Contract.ABI,
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

	anyOutput, err := contract.CustomERC20Contract.ABI.Unpack(erc20BalanceOfMethod, res.Ret)
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
