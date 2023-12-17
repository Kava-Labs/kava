package keeper_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	tmprototypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cometbft/cometbft/crypto"
	tmtime "github.com/cometbft/cometbft/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/issuance/keeper"
	"github.com/kava-labs/kava/x/issuance/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite

	tApp       app.TestApp
	keeper     keeper.Keeper
	ctx        sdk.Context
	addrs      []string
	modAccount sdk.AccAddress
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()

	ctx := tApp.NewContext(true, tmprototypes.Header{Height: 1, Time: tmtime.Now()})
	tApp.InitializeFromGenesisStates()
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	var strAddrs []string
	for _, addr := range addrs {
		acc := tApp.GetAccountKeeper().NewAccountWithAddress(ctx, addr)
		tApp.GetAccountKeeper().SetAccount(ctx, acc)
		strAddrs = append(strAddrs, addr.String())
	}

	keeper := tApp.GetIssuanceKeeper()
	modAccount, err := sdk.AccAddressFromBech32("kava1cj7njkw2g9fqx4e768zc75dp9sks8u9znxrf0w")
	suite.Require().NoError(err)

	suite.tApp = tApp
	suite.ctx = ctx
	suite.keeper = keeper
	suite.addrs = strAddrs
	suite.modAccount = modAccount
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) authtypes.AccountI {
	ak := suite.tApp.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getBalance(addr sdk.AccAddress, denom string) sdk.Coin {
	bk := suite.tApp.GetBankKeeper()
	return bk.GetBalance(suite.ctx, addr, denom)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) authtypes.ModuleAccountI {
	sk := suite.tApp.GetAccountKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) TestGetSetParams() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(types.Params{Assets: []types.Asset(nil)}, params)
	asset := types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0)))
	params = types.NewParams([]types.Asset{asset})
	suite.keeper.SetParams(suite.ctx, params)
	newParams := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(params, newParams)
}

func (suite *KeeperTestSuite) TestIssueTokens() {
	type args struct {
		assets   []types.Asset
		sender   string
		tokens   sdk.Coin
		receiver string
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[2],
				tokens:   sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("othertoken", sdkmath.NewInt(100000)),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
				receiver: suite.modAccount.String(),
			},
			errArgs{
				expectPass: false,
				contains:   "cannot issue tokens to module account",
			},
		},
		{
			"paused issuance",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, true, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:   suite.addrs[0],
				tokens:   sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
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
			sender, _ := sdk.AccAddressFromBech32(tc.args.sender)
			receiver, _ := sdk.AccAddressFromBech32(tc.args.receiver)
			err := suite.keeper.IssueTokens(suite.ctx, tc.args.tokens, sender, receiver)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(sdk.NewCoins(tc.args.tokens), sdk.NewCoins(suite.getBalance(receiver, tc.args.tokens.Denom)))
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestIssueTokensRateLimited() {
	type args struct {
		assets    []types.Asset
		supplies  []types.AssetSupply
		sender    string
		tokens    sdk.Coin
		receiver  string
		blockTime time.Time
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(true, sdkmath.NewInt(10000000000), time.Hour*24)),
				},
				supplies: []types.AssetSupply{
					types.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour),
				},
				sender:    suite.addrs[0],
				tokens:    sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
				receiver:  suite.addrs[2],
				blockTime: suite.ctx.BlockTime().Add(time.Hour),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"over-limit issuance",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(true, sdkmath.NewInt(10000000000), time.Hour*24)),
				},
				supplies: []types.AssetSupply{
					types.NewAssetSupply(sdk.NewCoin("usdtoken", sdk.ZeroInt()), time.Hour),
				},
				sender:    suite.addrs[0],
				tokens:    sdk.NewCoin("usdtoken", sdkmath.NewInt(10000000001)),
				receiver:  suite.addrs[2],
				blockTime: suite.ctx.BlockTime().Add(time.Hour),
			},
			errArgs{
				expectPass: false,
				contains:   "asset supply over limit",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)
			for _, supply := range tc.args.supplies {
				suite.keeper.SetAssetSupply(suite.ctx, supply, supply.GetDenom())
			}
			suite.ctx = suite.ctx.WithBlockTime(tc.args.blockTime)
			sender, _ := sdk.AccAddressFromBech32(tc.args.sender)
			receiver, _ := sdk.AccAddressFromBech32(tc.args.receiver)
			err := suite.keeper.IssueTokens(suite.ctx, tc.args.tokens, sender, receiver)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().Equal(sdk.NewCoins(tc.args.tokens), sdk.NewCoins(suite.getBalance(receiver, tc.args.tokens.Denom)))
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRedeemTokens() {
	type args struct {
		assets        []types.Asset
		sender        string
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid denom redemption",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("othertoken", sdkmath.NewInt(100000)),
			},
			errArgs{
				expectPass: false,
				contains:   "",
			},
		},
		{
			"non-owner redemption",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[2],
				initialTokens: sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
			},
			errArgs{
				expectPass: false,
				contains:   "account not authorized",
			},
		},
		{
			"paused redemption",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, true, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
			},
			errArgs{
				expectPass: false,
				contains:   "asset is paused",
			},
		},
		{
			"redeem amount greater than balance",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:        suite.addrs[0],
				initialTokens: sdk.NewCoin("usdtoken", sdkmath.NewInt(100000)),
				redeemTokens:  sdk.NewCoin("usdtoken", sdkmath.NewInt(200000)),
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
			sk := suite.tApp.GetBankKeeper()
			err := sk.MintCoins(suite.ctx, types.ModuleAccountName, sdk.NewCoins(tc.args.initialTokens))
			suite.Require().NoError(err)
			sender, _ := sdk.AccAddressFromBech32(tc.args.sender)
			err = sk.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleAccountName, sender, sdk.NewCoins(tc.args.initialTokens))
			suite.Require().NoError(err)

			err = suite.keeper.RedeemTokens(suite.ctx, tc.args.redeemTokens, sender)

			if tc.errArgs.expectPass {
				suite.Require().NoError(err)
				initialSupply := sdk.NewCoins(tc.args.redeemTokens)
				moduleAccount := suite.getModuleAccount(types.ModuleAccountName)
				suite.Require().Equal(sdk.NewCoins(initialSupply.Sub(tc.args.redeemTokens)...), sdk.NewCoins(suite.getBalance(moduleAccount.GetAddress(), tc.args.redeemTokens.Denom)))
			} else {
				suite.Require().Error(err)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBlockAddress() {
	type args struct {
		assets      []types.Asset
		sender      string
		blockedAddr string
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
			"unblockable token",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, false, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				blockedAddr: suite.addrs[1],
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: false,
				contains:   "asset does not support block/unblock functionality",
			},
		},
		{
			"non-owner block",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
		{
			"block non-existing account",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				sender:      suite.addrs[0],
				blockedAddr: sdk.AccAddress(crypto.AddressHash([]byte("RandomAddr"))).String(),
				denom:       "usdtoken",
			},
			errArgs{
				expectPass: false,
				contains:   "cannot block account that does not exist in state",
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			params := types.NewParams(tc.args.assets)
			suite.keeper.SetParams(suite.ctx, params)
			sender, _ := sdk.AccAddressFromBech32(tc.args.sender)
			blockedAddr, _ := sdk.AccAddressFromBech32(tc.args.blockedAddr)
			err := suite.keeper.BlockAddress(suite.ctx, tc.args.denom, sender, blockedAddr)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				asset, found := suite.keeper.GetAsset(suite.ctx, tc.args.denom)
				blocked := false
				suite.Require().True(found)
				for _, blockedAddr := range asset.BlockedAddresses {
					if blockedAddr == tc.args.blockedAddr {
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
		assets      []types.Asset
		sender      string
		blockedAddr string
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{suite.addrs[1]}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
			sender, _ := sdk.AccAddressFromBech32(tc.args.sender)
			blockedAddr, _ := sdk.AccAddressFromBech32(tc.args.blockedAddr)
			err := suite.keeper.UnblockAddress(suite.ctx, tc.args.denom, sender, blockedAddr)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				asset, found := suite.keeper.GetAsset(suite.ctx, tc.args.denom)
				blocked := false
				suite.Require().True(found)
				for _, blockedAddr := range asset.BlockedAddresses {
					if blockedAddr == tc.args.blockedAddr {
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
		assets      []types.Asset
		sender      string
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, true, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
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

			sender, _ := sdk.AccAddressFromBech32(tc.args.sender)
			err := suite.keeper.SetPauseStatus(suite.ctx, sender, tc.args.denom, tc.args.endStatus)
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
		assets       []types.Asset
		initialCoins sdk.Coin
		blockedAddrs []string
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
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				initialCoins: sdk.NewCoin("usdtoken", sdkmath.NewInt(100000000)),
				denom:        "usdtoken",
				blockedAddrs: []string{suite.addrs[1], suite.addrs[2]},
			},
			errArgs{
				expectPass: true,
				contains:   "",
			},
		},
		{
			"invalid denom seize",
			args{
				assets: []types.Asset{
					types.NewAsset(suite.addrs[0], "usdtoken", []string{}, false, true, types.NewRateLimit(false, sdk.ZeroInt(), time.Duration(0))),
				},
				initialCoins: sdk.NewCoin("usdtoken", sdkmath.NewInt(100000000)),
				denom:        "othertoken",
				blockedAddrs: []string{suite.addrs[1], suite.addrs[2]},
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
			assetsWithBlockedAddrs := []types.Asset{}
			for _, asset := range tc.args.assets {
				asset.BlockedAddresses = tc.args.blockedAddrs
				assetsWithBlockedAddrs = append(assetsWithBlockedAddrs, asset)
			}
			params := types.NewParams(assetsWithBlockedAddrs)
			suite.keeper.SetParams(suite.ctx, params)
			sk := suite.tApp.GetBankKeeper()
			for _, addrStr := range tc.args.blockedAddrs {
				addr, _ := sdk.AccAddressFromBech32(addrStr)
				err := sk.MintCoins(suite.ctx, types.ModuleAccountName, sdk.NewCoins(tc.args.initialCoins))
				suite.Require().NoError(err)
				err = sk.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleAccountName, addr, sdk.NewCoins(tc.args.initialCoins))
			}

			err := suite.keeper.SeizeCoinsFromBlockedAddresses(suite.ctx, tc.args.denom)
			if tc.errArgs.expectPass {
				suite.Require().NoError(err, tc.name)
				asset, found := suite.keeper.GetAsset(suite.ctx, tc.args.denom)
				suite.Require().True(found)
				owner, _ := sdk.AccAddressFromBech32(asset.Owner)
				ownerCoinAmount := tc.args.initialCoins.Amount.Mul(sdkmath.NewInt(int64(len(tc.args.blockedAddrs))))
				suite.Require().Equal(sdk.NewCoins(sdk.NewCoin(tc.args.denom, ownerCoinAmount)), sdk.NewCoins(suite.getBalance(owner, tc.args.initialCoins.Denom)))
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
