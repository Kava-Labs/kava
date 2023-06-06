package e2e_test

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/tests/util"
	committeetypes "github.com/kava-labs/kava/x/committee/types"
	evmutiltypes "github.com/kava-labs/kava/x/evmutil/types"
)

// TestUpgradeHandler can be used to run tests post-upgrade. If an upgrade is enabled, all tests
// are run against the upgraded chain. However, this file is a good place to consolidate all
// acceptance tests for a given set of upgrade handlers.
func (suite *IntegrationTestSuite) TestUpgradeHandler() {
	suite.SkipIfUpgradeDisabled()
	fmt.Println("An upgrade has run!")
	suite.True(true)

	// Thorough testing of the upgrade handler for v0.24 depends on:
	// - chain starting from v0.23 template
	// - funded account has ibc denom for ATOM
	// - Stability committee existing with committee id 1

	// Uncomment & use these contexts to compare chain state before & after the upgrade occurs.
	beforeUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := util.CtxAtHeight(suite.UpgradeHeight)

	// check x/evmutil module consensus version has been updated
	suite.Run("x/evmutil consensus version 1 -> 2", func() {
		before, err := suite.Kava.Upgrade.ModuleVersions(
			beforeUpgradeCtx,
			&upgradetypes.QueryModuleVersionsRequest{
				ModuleName: evmutiltypes.ModuleName,
			},
		)
		suite.NoError(err)
		suite.Equal(uint64(1), before.ModuleVersions[0].Version)

		after, err := suite.Kava.Upgrade.ModuleVersions(
			afterUpgradeCtx,
			&upgradetypes.QueryModuleVersionsRequest{
				ModuleName: evmutiltypes.ModuleName,
			},
		)
		suite.NoError(err)
		suite.Equal(uint64(2), after.ModuleVersions[0].Version)
	})

	// check evmutil params before & after upgrade
	suite.Run("x/evmutil AllowedCosmosDenoms updated", func() {
		before, err := suite.Kava.Evmutil.Params(beforeUpgradeCtx, &evmutiltypes.QueryParamsRequest{})
		suite.NoError(err)
		suite.Len(before.Params.AllowedCosmosDenoms, 0)

		after, err := suite.Kava.Evmutil.Params(afterUpgradeCtx, &evmutiltypes.QueryParamsRequest{})
		suite.NoError(err)
		suite.Len(after.Params.AllowedCosmosDenoms, 1)
		tokenInfo := after.Params.AllowedCosmosDenoms[0]
		suite.Equal(app.MainnetAtomDenom, tokenInfo.CosmosDenom)
	})

	// check x/evm param for allowed eip712 messages
	// use of these messages is performed in e2e_convert_cosmos_coins_test.go
	suite.Run("EIP712 signing allowed for new messages", func() {
		before, err := suite.Kava.Evm.Params(
			beforeUpgradeCtx,
			&evmtypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		suite.NotContains(before.Params.EIP712AllowedMsgs, app.EIP712AllowedMsgConvertCosmosCoinToERC20)
		suite.NotContains(before.Params.EIP712AllowedMsgs, app.EIP712AllowedMsgConvertCosmosCoinFromERC20)

		after, err := suite.Kava.Evm.Params(
			afterUpgradeCtx,
			&evmtypes.QueryParamsRequest{},
		)
		suite.NoError(err)
		suite.Contains(after.Params.EIP712AllowedMsgs, app.EIP712AllowedMsgConvertCosmosCoinToERC20)
		suite.Contains(after.Params.EIP712AllowedMsgs, app.EIP712AllowedMsgConvertCosmosCoinFromERC20)
	})

	// check stability committee permissions were updated
	suite.Run("stability committee ParamsChangePermission adds AllowedCosmosDenoms", func() {
		before, err := suite.Kava.Committee.Committee(
			beforeUpgradeCtx,
			&committeetypes.QueryCommitteeRequest{
				CommitteeId: app.MainnetStabilityCommitteeId,
			},
		)
		suite.NoError(err)
		fmt.Println("BEFORE: ", before.Committee)
		suite.NotContains(
			suite.getParamsChangePerm(before.Committee),
			app.AllowedParamsChangeAllowedCosmosDenoms,
		)

		after, err := suite.Kava.Committee.Committee(
			afterUpgradeCtx,
			&committeetypes.QueryCommitteeRequest{
				CommitteeId: app.MainnetStabilityCommitteeId,
			},
		)
		suite.NoError(err)
		fmt.Println("AFTER: ", after.Committee)
		suite.Contains(
			suite.getParamsChangePerm(after.Committee),
			app.AllowedParamsChangeAllowedCosmosDenoms,
		)
	})
}

func (suite *IntegrationTestSuite) getParamsChangePerm(anyComm *codectypes.Any) []committeetypes.AllowedParamsChange {
	var committee committeetypes.Committee
	err := suite.Kava.EncodingConfig.Marshaler.UnpackAny(anyComm, &committee)
	if err != nil {
		panic(err)
	}
	permissions := committee.GetPermissions()
	for _, perm := range permissions {
		if paramsChangePerm, ok := perm.(*committeetypes.ParamsChangePermission); ok {
			return paramsChangePerm.AllowedParamsChanges
		}
	}
	panic("no ParamsChangePermission found for stability committee")
}
