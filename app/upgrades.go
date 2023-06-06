package app

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	committeekeeper "github.com/kava-labs/kava/x/committee/keeper"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	evmutilkeeper "github.com/kava-labs/kava/x/evmutil/keeper"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

const (
	MainnetUpgradeName = "v0.24.0"
	TestnetUpgradeName = "v0.24.0-alpha.0"

	MainnetAtomDenom = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	TestnetHardDenom = "hard"

	MainnetStabilityCommitteeId = uint64(1)
	TestnetStabilityCommitteeId = uint64(1)
)

var (
	// Committee permission for changing AllowedCosmosDenoms param
	AllowedParamsChangeAllowedCosmosDenoms = committeetypes.AllowedParamsChange{
		Subspace: evmutiltypes.ModuleName,
		Key:      "AllowedCosmosDenoms",
	}

	// EIP712 allowed message for MsgConvertCosmosCoinToERC20
	EIP712AllowedMsgConvertCosmosCoinToERC20 = evmtypes.EIP712AllowedMsg{
		MsgTypeUrl:       "/kava.evmutil.v1beta1.MsgConvertCosmosCoinToERC20",
		MsgValueTypeName: "MsgConvertCosmosCoinToERC20",
		ValueTypes: []evmtypes.EIP712MsgAttrType{
			{
				Name: "initiator",
				Type: "string",
			},
			{
				Name: "receiver",
				Type: "string",
			},
			{
				Name: "amount",
				Type: "Coin",
			},
		},
		NestedTypes: nil,
	}
	// EIP712 allowed message for MsgConvertCosmosCoinFromERC20
	EIP712AllowedMsgConvertCosmosCoinFromERC20 = evmtypes.EIP712AllowedMsg{
		MsgTypeUrl:       "/kava.evmutil.v1beta1.MsgConvertCosmosCoinFromERC20",
		MsgValueTypeName: "MsgConvertCosmosCoinFromERC20",
		ValueTypes: []evmtypes.EIP712MsgAttrType{
			{
				Name: "initiator",
				Type: "string",
			},
			{
				Name: "receiver",
				Type: "string",
			},
			{
				Name: "amount",
				Type: "Coin",
			},
		},
		NestedTypes: nil,
	}
)

func (app App) RegisterUpgradeHandlers() {
	// register upgrade handler for mainnet
	app.upgradeKeeper.SetUpgradeHandler(MainnetUpgradeName, MainnetUpgradeHandler(app))

	// register upgrade handler for testnet
	app.upgradeKeeper.SetUpgradeHandler(TestnetUpgradeName, TestnetUpgradeHandler(app))

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	doUpgrade := upgradeInfo.Name == MainnetUpgradeName || upgradeInfo.Name == TestnetUpgradeName
	if doUpgrade && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func MainnetUpgradeHandler(app App) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("running mainnet upgrade handler")

		toVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return toVM, err
		}

		app.Logger().Info("initializing allowed_cosmos_denoms param of x/evmutil")
		allowedDenoms := []evmutiltypes.AllowedCosmosCoinERC20Token{
			{
				CosmosDenom: MainnetAtomDenom,
				// erc20 contract metadata
				Name:     "ATOM",
				Symbol:   "ATOM",
				Decimals: 6,
			},
		}
		InitializeEvmutilAllowedCosmosDenoms(ctx, &app.evmutilKeeper, allowedDenoms)

		app.Logger().Info("allowing cosmos coin conversion messaged in EIP712 signing")
		AllowEip712SigningForConvertMessages(ctx, app.evmKeeper)

		app.Logger().Info("allowing stability committee to update x/evmutil AllowedCosmosDenoms param")
		AddAllowedCosmosDenomsParamChangeToStabilityCommittee(
			ctx,
			app.interfaceRegistry,
			&app.committeeKeeper,
			MainnetStabilityCommitteeId,
		)

		return toVM, nil
	}
}

func TestnetUpgradeHandler(app App) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("running testnet upgrade handler")

		toVM, err := app.mm.RunMigrations(ctx, app.configurator, fromVM)
		if err != nil {
			return toVM, err
		}

		app.Logger().Info("initializing allowed_cosmos_denoms param of x/evmutil")
		// on testnet, IBC is not enabled. we initialize HARD tokens for conversion to EVM.
		allowedDenoms := []evmutiltypes.AllowedCosmosCoinERC20Token{
			{
				CosmosDenom: TestnetHardDenom,
				// erc20 contract metadata
				Name:     "HARD",
				Symbol:   "HARD",
				Decimals: 6,
			},
		}
		InitializeEvmutilAllowedCosmosDenoms(ctx, &app.evmutilKeeper, allowedDenoms)

		app.Logger().Info("allowing cosmos coin conversion messaged in EIP712 signing")
		AllowEip712SigningForConvertMessages(ctx, app.evmKeeper)

		app.Logger().Info("allowing stability committee to update x/evmutil AllowedCosmosDenoms param")
		AddAllowedCosmosDenomsParamChangeToStabilityCommittee(
			ctx,
			app.interfaceRegistry,
			&app.committeeKeeper,
			TestnetStabilityCommitteeId,
		)

		return toVM, nil
	}
}

// InitializeEvmutilAllowedCosmosDenoms sets the AllowedCosmosDenoms parameter of the x/evmutil module.
// This new parameter controls what cosmos denoms are allowed to be converted to ERC20 tokens.
func InitializeEvmutilAllowedCosmosDenoms(
	ctx sdk.Context,
	evmutilKeeper *evmutilkeeper.Keeper,
	allowedCoins []evmutiltypes.AllowedCosmosCoinERC20Token,
) {
	params := evmutilKeeper.GetParams(ctx)
	params.AllowedCosmosDenoms = allowedCoins
	if err := params.Validate(); err != nil {
		panic(fmt.Sprintf("x/evmutil params are not valid: %s", err))
	}
	evmutilKeeper.SetParams(ctx, params)
}

// AllowEip712SigningForConvertMessages adds the cosmos coin conversion messages to the
// allowed message types for EIP712 signing.
// The newly allowed messages are:
// - MsgConvertCosmosCoinToERC20
// - MsgConvertCosmosCoinFromERC20
func AllowEip712SigningForConvertMessages(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) {
	params := evmKeeper.GetParams(ctx)
	params.EIP712AllowedMsgs = append(
		params.EIP712AllowedMsgs,
		EIP712AllowedMsgConvertCosmosCoinToERC20,
		EIP712AllowedMsgConvertCosmosCoinFromERC20,
	)
	if err := params.Validate(); err != nil {
		panic(fmt.Sprintf("x/evm params are not valid: %s", err))
	}
	evmKeeper.SetParams(ctx, params)
}

// AddAllowedCosmosDenomsParamChangeToStabilityCommittee enables the stability committee
// to update the AllowedCosmosDenoms parameter of x/evmutil.
func AddAllowedCosmosDenomsParamChangeToStabilityCommittee(
	ctx sdk.Context,
	cdc codectypes.InterfaceRegistry,
	committeeKeeper *committeekeeper.Keeper,
	committeeId uint64,
) {
	// get committee
	committee, foundCommittee := committeeKeeper.GetCommittee(ctx, committeeId)
	if !foundCommittee {
		panic(fmt.Sprintf("expected to find committee with id %d but found none", committeeId))
	}

	permissions := committee.GetPermissions()

	// find & update the ParamsChangePermission
	foundPermission := false
	for i, permission := range permissions {
		if paramsChangePermission, ok := permission.(*committeetypes.ParamsChangePermission); ok {
			foundPermission = true
			paramsChangePermission.AllowedParamsChanges = append(
				paramsChangePermission.AllowedParamsChanges,
				AllowedParamsChangeAllowedCosmosDenoms,
			)
			permissions[i] = paramsChangePermission
			break
		}
	}

	// error if permission was not found & updated
	if !foundPermission {
		panic(fmt.Sprintf("no ParamsChangePermission found on committee with id %d", committeeId))
	}

	// update permissions
	committee.SetPermissions(permissions)
	if err := committee.Validate(); err != nil {
		panic(fmt.Sprintf("stability committee (id=%d) is invalid: %s", committeeId, err))
	}

	// save permission changes
	committeeKeeper.SetCommittee(ctx, committee)
}
