package keeper_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/x/kavamint/keeper"
	"github.com/kava-labs/kava/x/kavamint/testutil"
	"github.com/kava-labs/kava/x/kavamint/types"
)

const (
	custom = "custom"
)

type querierTestSuite struct {
	testutil.KavamintTestSuite

	legacyAmino *codec.LegacyAmino
	querier     sdk.Querier
}

func (suite *querierTestSuite) SetupTest() {
	suite.KavamintTestSuite.SetupTest()
	suite.legacyAmino = suite.App.LegacyAmino()
	suite.querier = keeper.NewQuerier(suite.Keeper, suite.legacyAmino)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(querierTestSuite))
}

func (suite *querierTestSuite) assertQuerierResponse(expected interface{}, actual []byte) {
	expectedJson, err := suite.legacyAmino.MarshalJSONIndent(expected, "", "  ")
	suite.Require().NoError(err)
	suite.Require().Equal(string(expectedJson), string(actual))
}

func (suite *querierTestSuite) Test_Querier_UnknownPath() {
	suite.SetupTest()

	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, "not-a-real-query"}, "/"),
		Data: nil,
	}

	_, err := suite.querier(suite.Ctx, []string{"not-a-real-query"}, query)
	suite.Error(err)
}

func (suite *querierTestSuite) Test_Querier() {
	testCases := []struct {
		name              string
		params            types.Params
		bondedRatio       sdk.Dec
		expectedInflation sdk.Dec
	}{
		{
			name: "basic inflation",
			params: types.NewParams(
				sdk.MustNewDecFromStr("0.42"),
				sdk.MustNewDecFromStr("0.20"),
			),
			bondedRatio: sdk.MustNewDecFromStr("0.5"),
			// should be community + 0.5*staking
			expectedInflation: sdk.MustNewDecFromStr("0.52"),
		},
		{
			name: "100 percent bonded is simple addition of inflation",
			params: types.NewParams(
				sdk.MustNewDecFromStr("0.42"),
				sdk.MustNewDecFromStr("0.20"),
			),
			bondedRatio:       sdk.OneDec(), // 100% bonded means inflation is just staking + community
			expectedInflation: sdk.MustNewDecFromStr("0.62"),
		},
		{
			name: "0 percent bonded is just community inflation",
			params: types.NewParams(
				sdk.MustNewDecFromStr("0.42"),
				sdk.MustNewDecFromStr("0.20"),
			),
			bondedRatio:       sdk.ZeroDec(), // 0% bonded means inflation is just community
			expectedInflation: sdk.MustNewDecFromStr("0.42"),
		},
		{
			name: "no inflation is no inflation",
			params: types.NewParams(
				sdk.ZeroDec(),
				sdk.ZeroDec(),
			),
			bondedRatio:       sdk.MustNewDecFromStr("0.3"),
			expectedInflation: sdk.ZeroDec(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			suite.Keeper.SetParams(suite.Ctx, tc.params)
			suite.SetBondedTokenRatio(tc.bondedRatio)
			staking.EndBlocker(suite.Ctx, suite.StakingKeeper)

			// query parameters
			query := abci.RequestQuery{
				Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryParameters}, "/"),
				Data: nil,
			}
			data, err := suite.querier(suite.Ctx, []string{types.QueryParameters}, query)
			suite.NoError(err)
			suite.NotNil(data)

			suite.assertQuerierResponse(tc.params, data)

			// query inflation
			query.Path = strings.Join([]string{custom, types.QuerierRoute, types.QueryInflation}, "/")
			data, err = suite.querier(suite.Ctx, []string{types.QueryInflation}, query)
			suite.NoError(err)
			suite.NotNil(data)

			suite.assertQuerierResponse(tc.expectedInflation, data)
		})
	}
}
