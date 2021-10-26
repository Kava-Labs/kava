package keeper_test

// import (
// 	gocontext "context"
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/suite"
// 	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

// 	"github.com/cosmos/cosmos-sdk/baseapp"
// 	"github.com/cosmos/cosmos-sdk/simapp"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/types/query"
// 	"github.com/cosmos/cosmos-sdk/x/distribution/types"
// 	"github.com/cosmos/cosmos-sdk/x/staking"
// 	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
// 	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
// )

// type KeeperTestSuite struct {
// 	suite.Suite

// 	app         *simapp.SimApp
// 	ctx         sdk.Context
// 	queryClient types.QueryClient
// 	addrs       []sdk.AccAddress
// 	valAddrs    []sdk.ValAddress
// }

// func (suite *KeeperTestSuite) SetupTest() {
// 	app := simapp.Setup(false)
// 	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

// 	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
// 	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
// 	queryClient := types.NewQueryClient(queryHelper)

// 	suite.app = app
// 	suite.ctx = ctx
// 	suite.queryClient = queryClient

// 	suite.addrs = simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000000000))
// 	suite.valAddrs = simapp.ConvertAddrsToValAddrs(suite.addrs)
// }

// func (suite *KeeperTestSuite) TestGRPCParams() {
// 	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

// 	var (
// 		params    types.Params
// 		req       *types.QueryParamsRequest
// 		expParams types.Params
// 	)

// 	testCases := []struct {
// 		msg      string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"empty params request",
// 			func() {
// 				req = &types.QueryParamsRequest{}
// 				expParams = types.DefaultParams()
// 			},
// 			true,
// 		},
// 		{
// 			"valid request",
// 			func() {
// 				params = types.Params{
// 					CommunityTax:        sdk.NewDecWithPrec(3, 1),
// 					BaseProposerReward:  sdk.NewDecWithPrec(2, 1),
// 					BonusProposerReward: sdk.NewDecWithPrec(1, 1),
// 					WithdrawAddrEnabled: true,
// 				}

// 				app.DistrKeeper.SetParams(ctx, params)
// 				req = &types.QueryParamsRequest{}
// 				expParams = params
// 			},
// 			true,
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", testCase.msg), func() {
// 			testCase.malleate()

// 			paramsRes, err := queryClient.Params(gocontext.Background(), req)

// 			if testCase.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(paramsRes)
// 				suite.Require().Equal(paramsRes.Params, expParams)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }
