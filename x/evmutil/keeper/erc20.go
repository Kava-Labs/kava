package keeper

import (
	"encoding/hex"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/evmutil/types"
)

const (
	erc20BalanceOfMethod = "balanceOf"
)

// DeployTestMintableERC20Contract deploys an ERC20 contract on the EVM as the
// module account and returns the address of the contract. This contract has
// minting permissions for the module account.
// Derived from tharsis/evmos
// https://github.com/tharsis/evmos/blob/ee54f496551df937915ff6f74a94732a35abc505/x/erc20/keeper/evm.go
func (k Keeper) DeployTestMintableERC20Contract(
	ctx sdk.Context,
	name string,
	symbol string,
	decimals uint8,
) (types.InternalEVMAddress, error) {
	ctorArgs, err := types.ERC20MintableBurnableContract.ABI.Pack(
		"", // Empty string for contract constructor
		name,
		symbol,
		decimals,
	)
	if err != nil {
		return types.InternalEVMAddress{}, errorsmod.Wrapf(err, "token %v is invalid", name)
	}

	data := make([]byte, len(types.ERC20MintableBurnableContract.Bin)+len(ctorArgs))
	copy(
		data[:len(types.ERC20MintableBurnableContract.Bin)],
		types.ERC20MintableBurnableContract.Bin,
	)
	copy(
		data[len(types.ERC20MintableBurnableContract.Bin):],
		ctorArgs,
	)

	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleEVMAddress.Bytes())
	if err != nil {
		return types.InternalEVMAddress{}, err
	}

	contractAddr := crypto.CreateAddress(types.ModuleEVMAddress, nonce)
	_, err = k.CallEVMWithData(ctx, types.ModuleEVMAddress, nil, data)
	if err != nil {
		return types.InternalEVMAddress{}, fmt.Errorf("failed to deploy ERC20 for %s: %w", name, err)
	}

	return types.NewInternalEVMAddress(contractAddr), nil
}

// DeployKavaWrappedCosmosCoinERC20Contract validates token details and then deploys an ERC20
// contract with the token metadata.
// This method does NOT check if a token for the provided SdkDenom has already been deployed.
func (k Keeper) DeployKavaWrappedCosmosCoinERC20Contract(
	ctx sdk.Context,
	token types.AllowedCosmosCoinERC20Token,
) (types.InternalEVMAddress, error) {
	if err := token.Validate(); err != nil {
		return types.InternalEVMAddress{}, errorsmod.Wrapf(err, "failed to deploy erc20 for sdk denom %s", token.CosmosDenom)
	}

	packedAbi, err := types.ERC20KavaWrappedCosmosCoinContract.ABI.Pack(
		"", // Empty string for contract constructor
		token.Name,
		token.Symbol,
		uint8(token.Decimals), // cast to uint8 is safe because of Validate()
	)
	if err != nil {
		return types.InternalEVMAddress{}, errorsmod.Wrapf(err, "failed to pack token with details %+v", token)
	}

	data := make([]byte, len(types.ERC20KavaWrappedCosmosCoinContract.Bin)+len(packedAbi))
	copy(
		data[:len(types.ERC20KavaWrappedCosmosCoinContract.Bin)],
		types.ERC20KavaWrappedCosmosCoinContract.Bin,
	)
	copy(
		data[len(types.ERC20KavaWrappedCosmosCoinContract.Bin):],
		packedAbi,
	)

	nonce, err := k.accountKeeper.GetSequence(ctx, types.ModuleEVMAddress.Bytes())
	if err != nil {
		return types.InternalEVMAddress{}, err
	}

	contractAddr := crypto.CreateAddress(types.ModuleEVMAddress, nonce)
	_, err = k.CallEVMWithData(ctx, types.ModuleEVMAddress, nil, data)
	if err != nil {
		return types.InternalEVMAddress{}, fmt.Errorf("failed to deploy ERC20 %s (nonce=%d, data=%s): %s", token.Name, nonce, hex.EncodeToString(data), err)
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
		types.ERC20MintableBurnableContract.ABI,
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
		types.ERC20MintableBurnableContract.ABI,
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

	anyOutput, err := types.ERC20MintableBurnableContract.ABI.Unpack(erc20BalanceOfMethod, res.Ret)
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
