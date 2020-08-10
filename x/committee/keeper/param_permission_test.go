package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	bep3types "github.com/kava-labs/kava/x/bep3/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/committee/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

type PermissionTestSuite struct {
	suite.Suite
	cdc *codec.Codec
}

func (suite *PermissionTestSuite) SetupTest() {
	app := app.NewTestApp()
	suite.cdc = app.Codec()
}

func (suite *PermissionTestSuite) TestSubParamChangePermission_Allows() {
	// cdp CollateralParams
	testCPs := cdptypes.CollateralParams{
		{
			Denom:               "bnb",
			LiquidationRatio:    d("2.0"),
			DebtLimit:           c("usdx", 1000000000000),
			StabilityFee:        d("1.000000001547125958"),
			LiquidationPenalty:  d("0.05"),
			AuctionSize:         i(100),
			Prefix:              0x20,
			ConversionFactor:    i(6),
			SpotMarketID:        "bnb:usd",
			LiquidationMarketID: "bnb:usd",
		},
		{
			Denom:               "btc",
			LiquidationRatio:    d("1.5"),
			DebtLimit:           c("usdx", 1000000000),
			StabilityFee:        d("1.000000001547125958"),
			LiquidationPenalty:  d("0.10"),
			AuctionSize:         i(1000),
			Prefix:              0x30,
			ConversionFactor:    i(8),
			SpotMarketID:        "btc:usd",
			LiquidationMarketID: "btc:usd",
		},
	}
	testCPUpdatedDebtLimit := make(cdptypes.CollateralParams, len(testCPs))
	copy(testCPUpdatedDebtLimit, testCPs)
	testCPUpdatedDebtLimit[0].DebtLimit = c("usdx", 5000000)

	// cdp DebtParam
	testDP := cdptypes.DebtParam{
		Denom:            "usdx",
		ReferenceAsset:   "usd",
		ConversionFactor: i(6),
		DebtFloor:        i(10000000),
		SavingsRate:      d("0.95"),
	}
	testDPUpdatedDebtFloor := testDP
	testDPUpdatedDebtFloor.DebtFloor = i(1000)

	// cdp Genesis
	testCDPParams := cdptypes.DefaultParams()
	testCDPParams.CollateralParams = testCPs
	testCDPParams.DebtParam = testDP
	testCDPParams.GlobalDebtLimit = testCPs[0].DebtLimit.Add(testCPs[0].DebtLimit) // correct global debt limit to pass genesis validation

	testDeputy, err := sdk.AccAddressFromBech32("kava1xy7hrjy9r0algz9w3gzm8u6mrpq97kwta747gj")
	suite.Require().NoError(err)
	// bep3 Asset Params
	testAPs := bep3types.AssetParams{
		bep3types.AssetParam{
			Denom:  "bnb",
			CoinID: 714,
			SupplyLimit: bep3types.SupplyLimit{
				Limit:          sdk.NewInt(350000000000000),
				TimeLimited:    false,
				TimeBasedLimit: sdk.ZeroInt(),
				TimePeriod:     time.Hour,
			},
			Active:        true,
			DeputyAddress: testDeputy,
			FixedFee:      sdk.NewInt(1000),
			MinSwapAmount: sdk.OneInt(),
			MaxSwapAmount: sdk.NewInt(1000000000000),
			MinBlockLock:  bep3types.DefaultMinBlockLock,
			MaxBlockLock:  bep3types.DefaultMaxBlockLock,
		},
		bep3types.AssetParam{
			Denom:  "inc",
			CoinID: 9999,
			SupplyLimit: bep3types.SupplyLimit{
				Limit:          sdk.NewInt(100000000000000),
				TimeLimited:    true,
				TimeBasedLimit: sdk.NewInt(50000000000),
				TimePeriod:     time.Hour,
			},
			Active:        false,
			DeputyAddress: testDeputy,
			FixedFee:      sdk.NewInt(1000),
			MinSwapAmount: sdk.OneInt(),
			MaxSwapAmount: sdk.NewInt(1000000000000),
			MinBlockLock:  bep3types.DefaultMinBlockLock,
			MaxBlockLock:  bep3types.DefaultMaxBlockLock,
		},
	}
	testAPsUpdatedActive := make(bep3types.AssetParams, len(testAPs))
	copy(testAPsUpdatedActive, testAPs)
	testAPsUpdatedActive[1].Active = true

	// bep3 Genesis
	testBep3Params := bep3types.DefaultParams()
	testBep3Params.AssetParams = testAPs

	// pricefeed Markets
	testMs := pricefeedtypes.Markets{
		{
			MarketID:   "bnb:usd",
			BaseAsset:  "bnb",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
		{
			MarketID:   "btc:usd",
			BaseAsset:  "btc",
			QuoteAsset: "usd",
			Oracles:    []sdk.AccAddress{},
			Active:     true,
		},
	}
	testMsUpdatedActive := make(pricefeedtypes.Markets, len(testMs))
	copy(testMsUpdatedActive, testMs)
	testMsUpdatedActive[1].Active = true

	testcases := []struct {
		name          string
		genState      []app.GenesisState
		permission    types.SubParamChangePermission
		pubProposal   types.PubProposal
		expectAllowed bool
	}{
		{
			name: "normal",
			genState: []app.GenesisState{
				newPricefeedGenState([]string{"bnb", "btc"}, []sdk.Dec{d("15.01"), d("9500")}),
				newCDPGenesisState(testCDPParams),
				newBep3GenesisState(testBep3Params),
			},
			permission: types.SubParamChangePermission{
				AllowedParams: types.AllowedParams{
					{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyDebtThreshold)},
					{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyCollateralParams)},
					{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyDebtParam)},
					{Subspace: bep3types.ModuleName, Key: string(bep3types.KeyAssetParams)},
					{Subspace: pricefeedtypes.ModuleName, Key: string(pricefeedtypes.KeyMarkets)},
				},
				AllowedCollateralParams: types.AllowedCollateralParams{
					{
						Denom:        "bnb",
						DebtLimit:    true,
						StabilityFee: true,
					},
					{ // TODO currently even if a perm doesn't allow a change in one element it must still be present in list
						Denom: "btc",
					},
				},
				AllowedDebtParam: types.AllowedDebtParam{
					DebtFloor: true,
				},
				AllowedAssetParams: types.AllowedAssetParams{
					{
						Denom: "bnb",
					},
					{
						Denom:  "inc",
						Active: true,
					},
				},
				AllowedMarkets: types.AllowedMarkets{
					{
						MarketID: "bnb:usd",
					},
					{
						MarketID: "btc:usd",
						Active:   true,
					},
				},
			},
			pubProposal: paramstypes.NewParameterChangeProposal(
				"A Title",
				"A description for this proposal.",
				[]paramstypes.ParamChange{
					{
						Subspace: cdptypes.ModuleName,
						Key:      string(cdptypes.KeyDebtThreshold),
						Value:    string(suite.cdc.MustMarshalJSON(i(1234))),
					},
					{
						Subspace: cdptypes.ModuleName,
						Key:      string(cdptypes.KeyCollateralParams),
						Value:    string(suite.cdc.MustMarshalJSON(testCPUpdatedDebtLimit)),
					},
					{
						Subspace: cdptypes.ModuleName,
						Key:      string(cdptypes.KeyDebtParam),
						Value:    string(suite.cdc.MustMarshalJSON(testDPUpdatedDebtFloor)),
					},
					{
						Subspace: bep3types.ModuleName,
						Key:      string(bep3types.KeyAssetParams),
						Value:    string(suite.cdc.MustMarshalJSON(testAPsUpdatedActive)),
					},
					{
						Subspace: pricefeedtypes.ModuleName,
						Key:      string(pricefeedtypes.KeyMarkets),
						Value:    string(suite.cdc.MustMarshalJSON(testMsUpdatedActive)),
					},
				},
			),
			expectAllowed: true,
		},
		{
			name:          "not allowed (wrong pubproposal type)",
			permission:    types.SubParamChangePermission{},
			pubProposal:   govtypes.NewTextProposal("A Title", "A description for this proposal."),
			expectAllowed: false,
		},
		{
			name:          "not allowed (nil pubproposal)",
			permission:    types.SubParamChangePermission{},
			pubProposal:   nil,
			expectAllowed: false,
		},
		// TODO more cases
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			tApp := app.NewTestApp()
			ctx := tApp.NewContext(true, abci.Header{})
			tApp.InitializeFromGenesisStates(tc.genState...)

			suite.Equal(
				tc.expectAllowed,
				tc.permission.Allows(ctx, tApp.Codec(), tApp.GetParamsKeeper(), tc.pubProposal),
			)
		})
	}

}
func TestPermissionTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionTestSuite))
}
