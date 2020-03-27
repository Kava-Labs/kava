package keeper_test
import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	addr, _ := sdk.AccAddressFromBech32("kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw")
	claimQueryParams := types.NewQueryClaimsParams(addr, "bnb")
	query := abci.RequestQuery{
		Path: strings.Join([]string{"custom", types.QuerierRoute, types.QueryGetClaims}, "/"),
		Data: types.ModuleCdc.MustMarshalJSON(claimQueryParams),
	}
	bz, err = querier(suite.ctx, []string{types.QueryGetClaims}, query)

	var claims types.Claims
	suite.Nil(types.ModuleCdc.UnmarshalJSON(bz, &claims))
	suite.Equal(1, len(claims))
}