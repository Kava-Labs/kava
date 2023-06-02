package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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
		expectedContracts := []types.DeployedCosmosCoinContract{
			contracts[2],
			contracts[4],
			contracts[0],
		}
		res, err := suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: denoms},
		)
		suite.NoError(err)
		// equal because it respects requested order
		suite.Equal(expectedContracts, res.DeployedCosmosCoinContracts)
	})

	suite.Run("un-deployed denoms get omitted from results", func() {
		denoms := []string{"doesnt-exist", ibcDenom, "also-doesnt-exist", "another-denom", "magic"}
		expectedContracts := []types.DeployedCosmosCoinContract{
			contracts[2],
			contracts[4],
			contracts[0],
		}
		res, err := suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: denoms},
		)
		suite.NoError(err)
		// equal because it respects requested order
		suite.Equal(expectedContracts, res.DeployedCosmosCoinContracts)
	})

	suite.Run("manages pagination of >100 denoms when requesting all", func() {
		// register 100 denoms
		for i := 1; i <= 100; i++ {
			suite.Keeper.SetDeployedCosmosCoinContract(
				suite.Ctx,
				fmt.Sprintf("denom-%d", i),
				testutil.RandomInternalEVMAddress(),
			)
		}

		// first page has 100 results
		res, err := suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{},
		)
		suite.NoError(err)
		// equal because it respects requested order
		suite.Len(res.DeployedCosmosCoinContracts, 100)
		fmt.Println(res.Pagination)
		suite.NotNil(res.Pagination.NextKey)

		// 2nd page has the rest
		res, err = suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{
				CosmosDenoms: []string{},
				Pagination: &query.PageRequest{
					Key: res.Pagination.NextKey,
				},
			},
		)
		suite.NoError(err)
		// equal because it respects requested order
		suite.Len(res.DeployedCosmosCoinContracts, len(contracts), "incorrect page 2 length")
		suite.Nil(res.Pagination.NextKey)
	})

	suite.Run("rejects requests for >100 denoms", func() {
		denoms := make([]string, 0, 101)
		for i := 1; i <= 100; i++ {
			denoms = append(denoms, fmt.Sprintf("nonexistent-%d", i))
		}

		// accepts 100 denoms
		res, err := suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: denoms},
		)
		suite.NoError(err)
		suite.Len(res.DeployedCosmosCoinContracts, 0)

		// rejects 101
		denoms = append(denoms, "nonexistent-101")
		_, err = suite.QueryClient.DeployedCosmosCoinContracts(
			context.Background(),
			&types.QueryDeployedCosmosCoinContractsRequest{CosmosDenoms: denoms},
		)
		suite.ErrorContains(err, "maximum of 100 denoms allowed per request")
	})
}
