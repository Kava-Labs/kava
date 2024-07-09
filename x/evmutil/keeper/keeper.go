package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// Keeper of the evmutil store.
// This keeper stores additional data related to evm accounts.
type Keeper struct {
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
	paramSubspace paramtypes.Subspace
	bankKeeper    types.BankKeeper
	evmKeeper     types.EvmKeeper
	accountKeeper types.AccountKeeper
}

// NewKeeper creates an evmutil keeper.
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	params paramtypes.Subspace,
	bk types.BankKeeper,
	ak types.AccountKeeper,
) Keeper {
	if !params.HasKeyTable() {
		params = params.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSubspace: params,
		bankKeeper:    bk,
		accountKeeper: ak,
	}
}

func (k *Keeper) SetEvmKeeper(evmKeeper types.EvmKeeper) {
	k.evmKeeper = evmKeeper
}

// SetDeployedCosmosCoinContract stores a single deployed ERC20KavaWrappedCosmosCoin contract address
func (k *Keeper) SetDeployedCosmosCoinContract(ctx sdk.Context, cosmosDenom string, contractAddress types.InternalEVMAddress) error {
	if err := sdk.ValidateDenom(cosmosDenom); err != nil {
		return errorsmod.Wrap(types.ErrInvalidCosmosDenom, cosmosDenom)
	}
	if contractAddress.IsNil() {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"attempting to register empty contract address for denom '%s'",
			cosmosDenom,
		)
	}
	store := ctx.KVStore(k.storeKey)
	storeKey := types.DeployedCosmosCoinContractKey(cosmosDenom)

	store.Set(storeKey, contractAddress.Bytes())
	return nil
}

// SetDeployedCosmosCoinContract gets a deployed ERC20KavaWrappedCosmosCoin contract address by cosmos denom
// Returns the stored address and a bool indicating if it was found or not
func (k *Keeper) GetDeployedCosmosCoinContract(ctx sdk.Context, cosmosDenom string) (types.InternalEVMAddress, bool) {
	store := ctx.KVStore(k.storeKey)
	storeKey := types.DeployedCosmosCoinContractKey(cosmosDenom)
	bz := store.Get(storeKey)
	found := len(bz) != 0
	return types.BytesToInternalEVMAddress(bz), found
}

// IterateAllDeployedCosmosCoinContracts iterates through all the deployed ERC20 contracts representing
// cosmos-sdk coins. If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAllDeployedCosmosCoinContracts(ctx sdk.Context, cb func(types.DeployedCosmosCoinContract) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DeployedCosmosCoinContractKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		contract := types.NewDeployedCosmosCoinContract(
			types.DenomFromDeployedCosmosCoinContractKey(iterator.Key()),
			types.BytesToInternalEVMAddress(iterator.Value()),
		)
		if cb(contract) {
			break
		}
	}
}
