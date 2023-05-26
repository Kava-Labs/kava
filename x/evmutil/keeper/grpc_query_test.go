package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type grpcQueryTestSuite struct {
	suite.Suite

	App         app.TestApp
	Keeper      keeper.Keeper
	Ctx         sdk.Context
	QueryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.App = app.NewTestApp()
	suite.App.InitializeFromGenesisStates()

	suite.Keeper = suite.App.GetEvmutilKeeper()
	suite.Ctx = suite.App.NewContext(true, tmproto.Header{Height: 1})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.App.GetEvmutilKeeper()))
	suite.QueryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestQueryParams() {
	expectedParams := types.DefaultParams()
	expectedParams.AllowedCosmosDenoms = append(
		expectedParams.AllowedCosmosDenoms,
		types.NewAllowedCosmosCoinERC20Token("cosmos-denom", "Cosmos Coin", "COSMOS", 6),
	)
	expectedParams.EnabledConversionPairs = types.NewConversionPairs(
		types.NewConversionPair(testutil.RandomInternalEVMAddress(), "evm-denom"),
		types.NewConversionPair(testutil.RandomInternalEVMAddress(), "evm-denom2"),
	)
	suite.Keeper.SetParams(suite.Ctx, expectedParams)

	params, err := suite.QueryClient.Params(
		context.Background(),
		&types.QueryParamsRequest{},
	)
	suite.Require().NoError(err)
	suite.Require().Equal(expectedParams, params.Params)
}

func (suite *grpcQueryTestSuite) TestQueryDeployedCosmosCoinContracts() {
	suite.Run("returns no error when no contracts deployed", func() {
		res, err := suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{},
		)
		suite.NoError(err)
		suite.Len(res.DeployedCosmosCoinContracts, 0)
	})

	// setup some deployed contracts
	ibcDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	contracts := []types.DeployedCosmosCoinContract{
		types.NewDeployedCosmosCoinContract("magic", testutil.RandomInternalEVMAddress()),
		types.NewDeployedCosmosCoinContract("hard", testutil.RandomInternalEVMAddress()),
		types.NewDeployedCosmosCoinContract(ibcDenom, testutil.RandomInternalEVMAddress()),
		types.NewDeployedCosmosCoinContract("swap", testutil.RandomInternalEVMAddress()),
		types.NewDeployedCosmosCoinContract("another-denom", testutil.RandomInternalEVMAddress()),
	}
	for _, c := range contracts {
		suite.Keeper.SetDeployedCosmosCoinContract(suite.Ctx, c.CosmosDenom, *c.Address)
	}

	suite.Run("returns all deployed contract addresses", func() {
		res, err := suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{},
		)
		suite.NoError(err)
		suite.ElementsMatch(contracts, res.DeployedCosmosCoinContracts)
	})

	suite.Run("returns deployed contract addresses for requested denoms", func() {
		denoms := []string{ibcDenom, "another-denom", "magic"}
		expectedContracts := make([]types.DeployedCosmosCoinContract, 0, len(denoms))
		for _, d := range denoms {
			// inefficient but readable
			for _, c := range contracts {
				if c.CosmosDenom == d {
					expectedContracts = append(expectedContracts, c)
				}
			}
		}

		res, err := suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: denoms},
		)
		suite.NoError(err)
		// equal because it respects requested order
		suite.Equal(expectedContracts, res.DeployedCosmosCoinContracts)
	})

	// suite.Run("handles querying un-deployed denoms")
}
