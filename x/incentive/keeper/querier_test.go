package keeper_test

import (
	"strings"

	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestQuerier() {
	suite.addObjectsToStore()
	querier := keeper.NewQuerier(suite.keeper)
	bz, err := querier(suite.ctx, []string{types.QueryGetParams}, abci.RequestQuery{})
	suite.NoError(err)
	suite.NotNil(bz)

	var p types.Params
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &p))

	claimQueryParams := types.NewQueryClaimsParams(suite.addrs[0], "bnb")
	query := abci.RequestQuery{
		Path: strings.Join([]string{"custom", types.QuerierRoute, types.QueryGetClaims}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(claimQueryParams),
	}
	bz, err = querier(suite.ctx, []string{types.QueryGetClaims}, query)

	var claims types.Claims
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &claims))
	suite.Equal(1, len(claims))
	suite.Equal(types.Claims{types.NewClaim(suite.addrs[0], c("ukava", 1000000), "bnb", 1)}, claims)
}
