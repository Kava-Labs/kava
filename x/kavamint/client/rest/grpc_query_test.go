package rest_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/kava-labs/kava/x/kavamint/types"
)

type IntegrationTestSuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := network.DefaultConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	var mintData types.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[types.ModuleName], &mintData))

	// inflation := sdk.MustNewDecFromStr("1.0")
	// mintData.Minter.Inflation = inflation
	// mintData.Params.InflationMin = inflation
	// mintData.Params.InflationMax = inflation

	mintDataBz, err := cfg.Codec.MarshalJSON(&mintData)
	s.Require().NoError(err)
	genesisState[types.ModuleName] = mintDataBz
	cfg.GenesisState = genesisState

	s.cfg = cfg
	s.network = network.New(s.T(), cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryGRPC() {
	val := s.network.Validators[0]
	baseURL := val.APIAddress
	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		respType proto.Message
		expected proto.Message
	}{
		{
			"gRPC request params",
			fmt.Sprintf("%s/kava/kavamint/v1beta1/parameters", baseURL),
			map[string]string{},
			&types.QueryParamsResponse{},
			&types.QueryParamsResponse{
				Params: types.NewParams(sdk.NewDecWithPrec(100, 2), sdk.NewDecWithPrec(20, 2)),
			},
		},
		{
			"gRPC request inflation",
			fmt.Sprintf("%s/kava/kavamint/v1beta1/inflation", baseURL),
			map[string]string{},
			&types.QueryInflationResponse{},
			&types.QueryInflationResponse{
				Inflation: sdk.NewDec(1),
			},
		},
	}
	for _, tc := range testCases {
		resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
		s.Run(tc.name, func() {
			s.Require().NoError(err)
			s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
