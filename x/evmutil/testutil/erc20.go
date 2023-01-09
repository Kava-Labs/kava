package testutil

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/kava-labs/kava/x/evmutil/contract"
	"github.com/kava-labs/kava/x/evmutil/keeper"
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
func DeployTestERC20Contract(
	ctx sdk.Context,
	k keeper.Keeper,
	name string,
	symbol string,
) (types.InternalEVMAddress, error) {
	ctorArgs, err := contract.CustomERC20Contract.ABI.Pack(
		"", // Empty string for contract constructor
		name,
		symbol,
	)
	if err != nil {
		return types.InternalEVMAddress{}, sdkerrors.Wrapf(err, "token %v is invalid", name)
	}

	data := make([]byte, len(contract.CustomERC20Contract.Bin)+len(ctorArgs))
	copy(
		data[:len(contract.CustomERC20Contract.Bin)],
		contract.CustomERC20Contract.Bin,
	)
	copy(
		data[len(contract.CustomERC20Contract.Bin):],
		ctorArgs,
	)

	nonce, err := k.GetAccountKeeper().GetSequence(ctx, types.ModuleEVMAddress.Bytes())
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
