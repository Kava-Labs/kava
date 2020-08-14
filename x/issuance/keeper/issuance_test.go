package keeper_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/issuance/keeper"
	"github.com/kava-labs/kava/x/issuance/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite

	keeper     keeper.Keeper
	app        app.TestApp
	ctx        sdk.Context
	addrs      []sdk.AccAddress
	modAccount sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	keeper := tApp.GetIssuanceKeeper()
	modAccount, err := sdk.AccAddressFromBech32("kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w")
	suite.Require().NoError(err)
	suite.app = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = addrs
	suite.modAccount = modAccount
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authexported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) supplyexported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) TestGetSetParams() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(types.Params{Assets: types.Assets(nil)}, params)
	asset := types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0)))
	params = types.NewParams(types.Assets{asset})
	suite.keeper.SetParams(suite.ctx, params)
	newParams := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(params, newParams)
}

func (suite *KeeperTestSuite) TestIssueTokens() {
	type args struct {
		assets   types.Assets
		sender   sdk.AccAddress
		tokens   sdk.Coin
		receiver sdk.AccAddress
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"valid issuance",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				receiver: suite.addrs[2],
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"non-owner issuance",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[2],
				tokens:   sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				receiver: suite.addrs[3],
			},
			errArgs{
				expectPass: false,
				contains:   "account not authorized",
			},
		},
		{
			"invalid denom",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("othertoken", sdk.NewInt(100000)),
				receiver: suite.addrs[2],
			},
			errArgs{
				expectPass: false,
				contains:   "no asset with input denom found",
			},
		},
		{
			"issue to blocked address",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				receiver: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "account is blocked",
			},
		},
		{
			"issue to module account",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				receiver: suite.modAccount,
			},
			errArgs{
				expectPass: false,
				contains:   "cannot issue tokens to module account",
			},
		},
		{
			"paused issuance",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, true, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				receiver: suite.addrs[1],
			},
			errArgs{
				expectPass: false,
				contains:   "asset is paused",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)
			err := suite.keeper.IssueTokens(suite.ctx, tc.args.tokens, tc.args.sender, tc.args.receiver)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				receiverAccount := suite.getAccount(tc.args.receiver)
				suite.Require().Equal(sdk.NewCoins(tc.args.tokens), receiverAccount.GetCoins())
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRedeemTokens() {
	type args struct {
		assets        types.Assets
		sender        sdk.AccAddress
		initialTokens sdk.Coin
		redeemTokens  sdk.Coin
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"valid redemption",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid denom redemption",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("othertoken", sdk.NewInt(100000)),
			},
			errArgs{
				expectPass: false,
				contains:   "",
			},
		},
		{
			"non-owner redemption",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[2],
				initialTokens: sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
			},
			errArgs{
				expectPass: false,
				contains:   "account not authorized",
			},
		},
		{
			"paused redemption",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, true, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
			},
			errArgs{
				expectPass: false,
				contains:   "asset is paused",
			},
		},
		{
			"redeem amount greater than balance",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdk.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdk.NewInt(200000)),
			},
			errArgs{
				expectPass: false,
				contains:   "insufficient funds",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)
			sk := suite.app.GetSupplyKeeper()
			err := sk.MintCoins(suite.ctx, types.ModuleAccountName, sdk.NewCoins(tc.args.initialTokens))
			suite.Require().NoError(err)
			err = sk.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleAccountName, tc.args.sender, sdk.NewCoins(tc.args.initialTokens))
			suite.Require().NoError(err)

			err = suite.keeper.RedeemTokens(suite.ctx, tc.args.redeemTokens, tc.args.sender)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				initialSupply := sdk.NewCoins(tc.args.redeemTokens)
				moduleAccount := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(initialSupply.Sub(sdk.NewCoins(tc.args.redeemTokens)), moduleAccount.GetCoins())
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBlockAddress() {
	type args struct {
		assets      types.Assets
		sender      sdk.AccAddress
		blockedAddr sdk.AccAddress
		denom       string
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"valid block",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				blockedAddr: suite.addrs[1],
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"non-owner block",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[2],
				blockedAddr: suite.addrs[1],
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: false,
				contains:   "account not authorized",
			},
		},
		{
			"invalid denom block",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				blockedAddr: suite.addrs[1],
				denom:       "othertoken",
			},
			errArgs{
				expectPass: false,
				contains:   "no asset with input denom found",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)

			err := suite.keeper.BlockAddress(suite.ctx, tc.args.denom, tc.args.sender, tc.args.blockedAddr)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				asset, found := suite.keeper.GetAsset(suite.ctx, tc.args.denom)
				blocked := false
				suite.Require().True(found)
				for _, blockedAddr := range asset.BlockedAddresses {
					if blockedAddr.Equals(tc.args.blockedAddr) {
						blocked = true
					}
				}
				suite.Require().True(blocked)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestUnblockAddress() {
	type args struct {
		assets      types.Assets
		sender      sdk.AccAddress
		blockedAddr sdk.AccAddress
		denom       string
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"valid unblock",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				blockedAddr: suite.addrs[1],
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"non-owner unblock",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[2],
				blockedAddr: suite.addrs[1],
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: false,
				contains:   "account not authorized",
			},
		},
		{
			"invalid denom block",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				blockedAddr: suite.addrs[1],
				denom:       "othertoken",
			},
			errArgs{
				expectPass: false,
				contains:   "no asset with input denom found",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)

			err := suite.keeper.UnblockAddress(suite.ctx, tc.args.denom, tc.args.sender, tc.args.blockedAddr)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				asset, found := suite.keeper.GetAsset(suite.ctx, tc.args.denom)
				blocked := false
				suite.Require().True(found)
				for _, blockedAddr := range asset.BlockedAddresses {
					if blockedAddr.Equals(tc.args.blockedAddr) {
						blocked = true
					}
				}
				suite.Require().False(blocked)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChangePauseStatus() {
	type args struct {
		assets      types.Assets
		sender      sdk.AccAddress
		startStatus bool
		endStatus   bool
		denom       string
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"valid pause",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				startStatus: false,
				endStatus:   true,
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"valid unpause",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, true, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				startStatus: true,
				endStatus:   false,
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"non-owner pause",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[2],
				startStatus: false,
				endStatus:   true,
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: false,
				contains:   "account not authorized",
			},
		},
		{
			"invalid denom pause",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				startStatus: true,
				endStatus:   false,
				denom:       "othertoken",
			},
			errArgs{
				expectPass: false,
				contains:   "no asset with input denom found",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)

			err := suite.keeper.SetPauseStatus(suite.ctx, tc.args.sender, tc.args.denom, tc.args.endStatus)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				asset, found := suite.keeper.GetAsset(suite.ctx, tc.args.denom)
				suite.Require().True(found)
				suite.Require().Equal(tc.args.endStatus, asset.Paused)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSeizeCoinsFromBlockedAddress() {
	type args struct {
		assets       types.Assets
		initialCoins sdk.Coin
		blockedAddrs []sdk.AccAddress
		denom        string
	}
	type errArgs struct {
		expectPass bool
		contains   string
	}
	testCases := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{
			"valid seize",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				initialCoins: sdk.NewCoin("usdtoken", sdk.NewInt(100000000)),
				denom:        "usdtoken",
				blockedAddrs: []sdk.AccAddress{suite.addrs[1], suite.addrs[2]},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid denom seize",
			args{
				assets: types.Assets{
					types.NewAsset(suite.addrs[0], "usdtoken", []sdk.AccAddress{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				initialCoins: sdk.NewCoin("usdtoken", sdk.NewInt(100000000)),
				denom:        "othertoken",
				blockedAddrs: []sdk.AccAddress{suite.addrs[1], suite.addrs[2]},
			},
			errArgs{
				expectPass: false,
				contains:   "no asset with input denom found",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			assetsWithBlockedAddrs := types.Assets{}
			for _, asset := range tc.args.assets {
				asset.BlockedAddresses = tc.args.blockedAddrs
				assetsWithBlockedAddrs = append(assetsWithBlockedAddrs, asset)
			}
			params := types.NewParams(assetsWithBlockedAddrs)
			suite.keeper.SetParams(suite.ctx, params)
			sk := suite.app.GetSupplyKeeper()
			for _, addr := range tc.args.blockedAddrs {
				err := sk.MintCoins(suite.ctx, types.ModuleAccountName, sdk.NewCoins(tc.args.initialCoins))
				suite.Require().NoError(err)
				err = sk.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleAccountName, addr, sdk.NewCoins(tc.args.initialCoins))
			}

			err := suite.keeper.SeizeCoinsFromBlockedAddresses(suite.ctx, tc.args.denom)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				asset, found := suite.keeper.GetAsset(suite.ctx, tc.args.denom)
				suite.Require().True(found)
				ownerAccount := suite.getAccount(asset.Owner)
				ownerCoinAmount := tc.args.initialCoins.Amount.Mul(sdk.NewInt(int64(len(tc.args.blockedAddrs))))
				suite.Require().Equal(sdk.NewCoins(sdk.NewCoin(tc.args.denom, ownerCoinAmount)), ownerAccount.GetCoins())
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
