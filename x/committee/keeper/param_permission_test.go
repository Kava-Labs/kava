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
	testDP := cdptypes.DebtParam{
		Denom:            "usdx",
		ReferenceAsset:   "usd",
		ConversionFactor: i(6),
		DebtFloor:        i(10000000),
		SavingsRate:      d("0.95"),
	}
	testDPUpdatedDebtFloor := testDP
	testDPUpdatedDebtFloor.DebtFloor = i(1000)

	testCPUpdatedDebtLimit := make(cdptypes.CollateralParams, len(testCPs))
	copy(testCPUpdatedDebtLimit, testCPs)
	testCPUpdatedDebtLimit[0].DebtLimit = c("usdx", 5000000)

	testCDPParams := cdptypes.DefaultParams()
	testCDPParams.CollateralParams = testCPs
	testCDPParams.DebtParam = testDP
	testCDPParams.GlobalDebtLimit = testCPs[0].DebtLimit.Add(testCPs[0].DebtLimit) // correct global debt limit to pass genesis validation

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
			},
			permission: types.SubParamChangePermission{
				AllowedParams: types.AllowedParams{
					{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyDebtThreshold)},
					{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyCollateralParams)},
					{Subspace: cdptypes.ModuleName, Key: string(cdptypes.KeyDebtParam)},
				},
				AllowedCollateralParams: types.AllowedCollateralParams{
					types.AllowedCollateralParam{
						Denom:        "bnb",
						DebtLimit:    true,
						StabilityFee: true,
					},
					types.AllowedCollateralParam{ // TODO currently even if a perm doesn't allow a change in one element it must still be present in list
						Denom: "btc",
					},
				},
				AllowedDebtParam: types.AllowedDebtParam{
					DebtFloor: true,
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
