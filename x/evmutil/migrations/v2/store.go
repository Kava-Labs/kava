package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kava-labs/kava/x/evmutil/types"
)

// MigrateStore performs in-place migrations from kava 10 to kava 11.
// The migration includes:
//
// - Setting the evmutil params in the paramstore
func MigrateStore(ctx sdk.Context, ss paramtypes.Subspace) error {
	if !ss.HasKeyTable() {
		ss = ss.WithKeyTable(types.ParamKeyTable())
	}
	params := types.DefaultParams()

	// add multichain erc20 stablecoins
	usdcAddr, err := types.NewInternalEVMAddressFromString("0xfA9343C3897324496A05fC75abeD6bAC29f8A40f")
	if err != nil {
		return err
	}
	usdtAddr, err := types.NewInternalEVMAddressFromString("0xB44a9B6905aF7c801311e8F4E76932ee959c663C")
	if err != nil {
		return err
	}
	daiAddr, err := types.NewInternalEVMAddressFromString("0x765277EebeCA2e31912C9946eAe1021199B39C61")
	if err != nil {
		return err
	}
	params.EnabledConversionPairs = types.ConversionPairs{
		{Denom: "erc20/multichain/usdc", KavaERC20Address: usdcAddr.Bytes()},
		{Denom: "erc20/multichain/usdt", KavaERC20Address: usdtAddr.Bytes()},
		{Denom: "erc20/multichain/dai", KavaERC20Address: daiAddr.Bytes()},
	}

	ss.SetParamSet(ctx, &params)
	return nil
}
