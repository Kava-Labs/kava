package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type keeperTestSuite struct {
	testutil.Suite
}

func (suite *keeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *keeperTestSuite) TestDeployedCosmosCoinContractStoreState() {
	suite.Run("returns nil for nonexistent denom", func() {
		suite.SetupTest()
		addr, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, "undeployed-denom")
		suite.False(found)
		suite.Equal(addr, types.InternalEVMAddress{})
	})

	suite.Run("handles setting & getting a contract address", func() {
		suite.SetupTest()
		denom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
		address := testutil.RandomInternalEVMAddress()

		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, denom, address)
		suite.NoError(err)

		stored, found := suite.Keeper.GetDeployedCosmosCoinContract(suite.Ctx, denom)
		suite.True(found)
		suite.Equal(address, stored)
	})

	suite.Run("fails when setting an invalid denom", func() {
		suite.SetupTest()
		invalidDenom := ""
		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, invalidDenom, testutil.RandomInternalEVMAddress())
		suite.ErrorContains(err, "invalid cosmos denom")
	})

	suite.Run("fails when setting 0 address", func() {
		suite.SetupTest()
		invalidAddr := types.InternalEVMAddress{}
		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, "denom", invalidAddr)
		suite.ErrorContains(err, "attempting to register empty contract address")
	})
}

func (suite *keeperTestSuite) TestIterateAllDeployedCosmosCoinContracts() {
	suite.SetupTest()
	address := testutil.RandomInternalEVMAddress()
	denoms := []string{}
	register := func(denom string) {
		addr := testutil.RandomInternalEVMAddress()
		if denom == "waldo" {
			addr = address
		}
		err := suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, denom, addr)
		suite.NoError(err)
		denoms = append(denoms, denom)
	}

	// register some contracts
	register("magic")
	register("popcorn")
	register("waldo")
	register("zzz")
	register("ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2")

	suite.Run("stops when told", func() {
		// test out stopping the iteration
		// NOTE: don't actually look for a single contract this way. the keys are deterministic by denom.
		var contract types.DeployedCosmosCoinContract
		suite.Keeper.IterateAllDeployedCosmosCoinContracts(suite.Ctx, func(c types.DeployedCosmosCoinContract) bool {
			contract = c
			return c.CosmosDenom == "waldo"
		})
		suite.Equal(types.NewDeployedCosmosCoinContract("waldo", address), contract)
	})

	suite.Run("iterates all contracts", func() {
		foundDenoms := make([]string, 0, len(denoms))
		suite.Keeper.IterateAllDeployedCosmosCoinContracts(suite.Ctx, func(c types.DeployedCosmosCoinContract) bool {
			foundDenoms = append(foundDenoms, c.CosmosDenom)
			return false
		})
		suite.Len(foundDenoms, len(denoms))
		suite.ElementsMatch(denoms, foundDenoms)
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(keeperTestSuite))
}
