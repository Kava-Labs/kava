package keeper_test

import (
	"testing"

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
			Denom:              "bnb",
			LiquidationRatio:   d("2.0"),
			DebtLimit:          c("usdx", 1000000000000),
			StabilityFee:       d("1.000000001547125958"),
			LiquidationPenalty: d("0.05"),
			AuctionSize:        i(100),
			Prefix:             0x20,
			ConversionFactor:   i(6),
			MarketID:           "bnb:usd",
		},
		{
			Denom:              "btc",
			LiquidationRatio:   d("1.5"),
			DebtLimit:          c("usdx", 1000000000),
			StabilityFee:       d("1.000000001547125958"),
			LiquidationPenalty: d("0.10"),
			AuctionSize:        i(1000),
			Prefix:             0x30,
			ConversionFactor:   i(8),
			MarketID:           "btc:usd",
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

	// bep3 Asset Params
	testAPs := bep3types.AssetParams{
		{
			Denom:  "bnb",
			CoinID: 714,
			Limit:  i(100000000000),
			Active: true,
		},
		{
			Denom:  "inc",
			CoinID: 9999,
			Limit:  i(100),
			Active: false,
		},
	}
	testAPsUpdatedActive := make(bep3types.AssetParams, len(testAPs))
	copy(testAPsUpdatedActive, testAPs)
	testAPsUpdatedActive[1].Active = true

	// bep3 Genesis
	testBep3Params := bep3types.DefaultParams()
	testBep3Params.SupportedAssets = testAPs

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
					{Subspace: bep3types.ModuleName, Key: string(bep3types.KeySupportedAssets)},
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
						Key:      string(bep3types.KeySupportedAssets),
						Value:    string(suite.cdc.MustMarshalJSON(testAPsUpdatedActive)),
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
