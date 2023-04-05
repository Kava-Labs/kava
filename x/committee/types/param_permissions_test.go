package types_test

import (
	fmt "fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	types "github.com/kava-labs/kava/x/committee/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

type ParamsChangeTestSuite struct {
	suite.Suite

	ctx sdk.Context
	pk  types.ParamKeeper

	cdpCollateralParams       cdptypes.CollateralParams
	cdpDebtParam              cdptypes.DebtParam
	cdpCollateralRequirements []types.SubparamRequirement
}

func (suite *ParamsChangeTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	suite.ctx = ctx
	suite.pk = tApp.GetParamsKeeper()

	suite.cdpDebtParam = cdptypes.DebtParam{
		Denom:            "usdx",
		ReferenceAsset:   "usd",
		ConversionFactor: sdkmath.NewInt(6),
		DebtFloor:        sdkmath.NewInt(1000),
	}

	suite.cdpCollateralParams = cdptypes.CollateralParams{
		{
			Denom:                            "bnb",
			Type:                             "bnb-a",
			LiquidationRatio:                 sdk.MustNewDecFromStr("2.0"),
			DebtLimit:                        sdk.NewCoin("usdx", sdkmath.NewInt(100)),
			StabilityFee:                     sdk.MustNewDecFromStr("1.02"),
			LiquidationPenalty:               sdk.MustNewDecFromStr("0.05"),
			AuctionSize:                      sdkmath.NewInt(100),
			ConversionFactor:                 sdkmath.NewInt(6),
			SpotMarketID:                     "bnb:usd",
			LiquidationMarketID:              "bnb:usd",
			CheckCollateralizationIndexCount: sdkmath.NewInt(0),
		},
		{
			Denom:                            "btc",
			Type:                             "btc-a",
			LiquidationRatio:                 sdk.MustNewDecFromStr("1.5"),
			DebtLimit:                        sdk.NewCoin("usdx", sdkmath.NewInt(100)),
			StabilityFee:                     sdk.MustNewDecFromStr("1.01"),
			LiquidationPenalty:               sdk.MustNewDecFromStr("0.10"),
			AuctionSize:                      sdkmath.NewInt(1000),
			ConversionFactor:                 sdkmath.NewInt(8),
			SpotMarketID:                     "btc:usd",
			LiquidationMarketID:              "btc:usd",
			CheckCollateralizationIndexCount: sdkmath.NewInt(1),
			KeeperRewardPercentage:           sdk.MustNewDecFromStr("0.12"),
		},
	}
	suite.cdpCollateralRequirements = []types.SubparamRequirement{
		{
			Key:                        "type",
			Val:                        "bnb-a",
			AllowedSubparamAttrChanges: []string{"conversion_factor", "liquidation_ratio", "spot_market_id"},
		},
		{
			Key:                        "type",
			Val:                        "btc-a",
			AllowedSubparamAttrChanges: []string{"stability_fee", "debt_limit", "auction_size", "keeper_reward_percentage"},
		},
	}
}

func (s *ParamsChangeTestSuite) TestSingleSubparams_CdpDeptParams() {
	testcases := []struct {
		name        string
		expected    bool
		permission  types.AllowedParamsChange
		paramChange paramsproposal.ParamChange
	}{
		{
			name:     "allow changes to all allowed fields",
			expected: true,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"denom", "reference_asset", "conversion_factor", "debt_floor"},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "DebtParam",
				Value: `{
					"denom": "bnb",
					"reference_asset": "bnbx",
					"conversion_factor": "11",
					"debt_floor": "1200"
				}`,
			},
		},
		{
			name:     "allows changes only to certain fields",
			expected: true,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"denom", "debt_floor"},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "DebtParam",
				Value: `{
					"denom": "bnb",
					"reference_asset": "usd",
					"conversion_factor": "6",
					"debt_floor": "1100"
				}`,
			},
		},
		{
			name:     "fails if changing attr that is not allowed",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"denom", "debt_floor"},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "DebtParam",
				Value: `{
					"denom": "usdx",
					"reference_asset": "usd",
					"conversion_factor": "7",
					"debt_floor": "1000"
				}`,
			},
		},
		{
			name:     "fails if there are unexpected param change attrs",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"denom"},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "DebtParam",
				Value: `{
					"denom": "usdx",
					"reference_asset": "usd",
					"conversion_factor": "6",
					"debt_floor": "1000",
					"extra_attr": "123"
				}`,
			},
		},
		{
			name:     "fails if there are missing param change attrs",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"denom", "reference_asset"},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "DebtParam",
				// debt_floor is missing
				Value: `{
					"denom": "usdx",
					"reference_asset": "usd",
					"conversion_factor": "11.000000000000000000",
				}`,
			},
		},
		{
			name:     "fails if subspace does not match",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"denom"},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "auction",
				Key:      "DebtParam",
				Value: `{
					"denom": "usdx",
					"reference_asset": "usd",
					"conversion_factor": "6",
					"debt_floor": "1000"
				}`,
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			subspace, found := s.pk.GetSubspace(cdptypes.ModuleName)
			s.Require().True(found)
			subspace.Set(s.ctx, cdptypes.KeyDebtParam, s.cdpDebtParam)

			permission := types.ParamsChangePermission{
				AllowedParamsChanges: types.AllowedParamsChanges{tc.permission},
			}
			proposal := paramsproposal.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramsproposal.ParamChange{tc.paramChange},
			)
			s.Require().Equal(
				tc.expected,
				permission.Allows(s.ctx, s.pk, proposal),
			)
		})
	}
}

func (s *ParamsChangeTestSuite) TestMultiSubparams_CdpCollateralParams() {
	unchangedBnbValue := `{
		"denom": "bnb",
		"type": "bnb-a",
		"liquidation_ratio": "2.000000000000000000",
		"debt_limit": { "denom": "usdx", "amount": "100" },
		"stability_fee": "1.020000000000000000",
		"auction_size": "100",
		"liquidation_penalty": "0.050000000000000000",
		"spot_market_id": "bnb:usd",
		"liquidation_market_id": "bnb:usd",
		"keeper_reward_percentage": "0",
		"check_collateralization_index_count": "0",
		"conversion_factor": "6"
	}`
	unchangedBtcValue := `{
		"denom": "btc",
		"type": "btc-a",
		"liquidation_ratio": "1.500000000000000000",
		"debt_limit": { "denom": "usdx", "amount": "100" },
		"stability_fee": "1.010000000000000000",
		"auction_size": "1000",
		"liquidation_penalty": "0.100000000000000000",
		"spot_market_id": "btc:usd",
		"liquidation_market_id": "btc:usd",
		"keeper_reward_percentage": "0.12",
		"check_collateralization_index_count": "1",
		"conversion_factor": "8"
	}`

	testcases := []struct {
		name        string
		expected    bool
		permission  types.AllowedParamsChange
		paramChange paramsproposal.ParamChange
	}{
		{
			name:     "succeeds when changing allowed values and keeping not allowed the same",
			expected: true,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: s.cdpCollateralRequirements,
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				Value: `[{
					"denom": "bnb",
					"type": "bnb-a",
					"liquidation_ratio": "2.010000000000000000",
					"debt_limit": { "denom": "usdx", "amount": "100" },
					"stability_fee": "1.020000000000000000",
					"auction_size": "100",
					"liquidation_penalty": "0.050000000000000000",
					"spot_market_id": "bnbc:usd",
					"liquidation_market_id": "bnb:usd",
					"keeper_reward_percentage": "0",
					"check_collateralization_index_count": "0",
					"conversion_factor": "9"
				},
				{
					"denom": "btc",
					"type": "btc-a",
					"liquidation_ratio": "1.500000000000000000",
					"debt_limit": { "denom": "usdx", "amount": "200" },
					"stability_fee": "2.010000000000000000",
					"auction_size": "1200",
					"liquidation_penalty": "0.100000000000000000",
					"spot_market_id": "btc:usd",
					"liquidation_market_id": "btc:usd",
					"keeper_reward_percentage": "0.000000000000000000",
					"check_collateralization_index_count": "1",
					"conversion_factor": "8"
				}]`,
			},
		},
		{
			name:     "succeeds if nothing is changed",
			expected: true,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: s.cdpCollateralRequirements,
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				Value:    fmt.Sprintf("[%s, %s]", unchangedBnbValue, unchangedBtcValue),
			},
		},
		{
			name:     "fails if changed records are not the same length as existing records",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: s.cdpCollateralRequirements,
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				Value:    fmt.Sprintf("[%s]", unchangedBnbValue),
			},
		},
		{
			name:     "fails if incoming records are missing a existing record",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: s.cdpCollateralRequirements,
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				// same length as existing records but missing one with the correct key/value pair
				Value: fmt.Sprintf("[%s, %s]", unchangedBnbValue, unchangedBnbValue),
			},
		},
		{
			name:     "fails when changing an attribute that is not allowed",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: s.cdpCollateralRequirements,
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				// changed liquidation_ratio, which is not whitelisted
				Value: fmt.Sprintf("[%s, %s]", unchangedBnbValue, `{
					"denom": "btc",
					"type": "btc-a",
					"liquidation_ratio": "1.2",
					"debt_limit": { "denom": "usdx", "amount": "100" },
					"stability_fee": "1.01",
					"auction_size": "1000",
					"liquidation_penalty": "0.1",
					"spot_market_id": "btc:usd",
					"liquidation_market_id": "btc:usd",
					"keeper_reward_percentage": "0.12",
					"check_collateralization_index_count": "1",
					"conversion_factor": "8"
				}`),
			},
		},
		{
			name:     "fails when requirements does not include an existing record",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: []types.SubparamRequirement{s.cdpCollateralRequirements[0]},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				Value:    fmt.Sprintf("[%s, %s]", unchangedBnbValue, unchangedBtcValue),
			},
		},
		{
			name:     "fails when changes has missing key",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: []types.SubparamRequirement{s.cdpCollateralRequirements[0]},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				// missing check_collateralization_index_count
				Value: fmt.Sprintf("[%s, %s]", unchangedBnbValue, `{
					"denom": "btc",
					"type": "btc-a",
					"liquidation_ratio": "1.500000000000000000",
					"debt_limit": { "denom": "usdx", "amount": "100" },
					"stability_fee": "1.010000000000000000",
					"auction_size": "1000",
					"liquidation_penalty": "0.100000000000000000",
					"spot_market_id": "btc:usd",
					"liquidation_market_id": "btc:usd",
					"keeper_reward_percentage": "0.12",
					"conversion_factor": "8"
				}`),
			},
		},
		{
			name:     "fails when changes has same keys length but an unknown key",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: []types.SubparamRequirement{s.cdpCollateralRequirements[0]},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				// missspelled denom key
				Value: fmt.Sprintf("[%s, %s]", unchangedBnbValue, `{
					"denoms": "btc",
					"type": "btc-a",
					"liquidation_ratio": "1.500000000000000000",
					"debt_limit": { "denom": "usdx", "amount": "100" },
					"stability_fee": "1.010000000000000000",
					"auction_size": "1000",
					"liquidation_penalty": "0.100000000000000000",
					"spot_market_id": "btc:usd",
					"liquidation_market_id": "btc:usd",
					"keeper_reward_percentage": "0.12",
					"check_collateralization_index_count": "1",
					"conversion_factor": "8"
				}`),
			},
		},
		{
			name:     "fails when attr is not allowed and has different value",
			expected: false,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: []types.SubparamRequirement{s.cdpCollateralRequirements[0]},
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				// liquidation_ratio changed value but is not allowed
				Value: fmt.Sprintf("[%s, %s]", unchangedBnbValue, `{
					"denom": "btc",
					"type": "btc-a",
					"liquidation_ratio": "1.510000000000000000",
					"debt_limit": { "denom": "usdx", "amount": "100" },
					"stability_fee": "1.010000000000000000",
					"auction_size": "1000",
					"liquidation_penalty": "0.100000000000000000",
					"spot_market_id": "btc:usd",
					"liquidation_market_id": "btc:usd",
					"keeper_reward_percentage": "0.12",
					"check_collateralization_index_count": "1",
					"conversion_factor": "8"
				}`),
			},
		},
		{
			name:     "succeeds when param attr is not allowed but is same",
			expected: true,
			permission: types.AllowedParamsChange{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyCollateralParams),
				MultiSubparamsRequirements: s.cdpCollateralRequirements,
			},
			paramChange: paramsproposal.ParamChange{
				Subspace: "cdp",
				Key:      "CollateralParams",
				// liquidation_ratio is not allowed but the same
				// stability_fee is allowed but changed
				Value: fmt.Sprintf("[%s, %s]", unchangedBnbValue, `{
					"denom": "btc",
					"type": "btc-a",
					"liquidation_ratio": "1.500000000000000000",
					"debt_limit": { "denom": "usdx", "amount": "100" },
					"stability_fee": "1.020000000000000000",
					"auction_size": "1000",
					"liquidation_penalty": "0.100000000000000000",
					"spot_market_id": "btc:usd",
					"liquidation_market_id": "btc:usd",
					"keeper_reward_percentage": "0.12",
					"check_collateralization_index_count": "1",
					"conversion_factor": "8"
				}`),
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			subspace, found := s.pk.GetSubspace(cdptypes.ModuleName)
			s.Require().True(found)
			subspace.Set(s.ctx, cdptypes.KeyCollateralParams, s.cdpCollateralParams)

			permission := types.ParamsChangePermission{
				AllowedParamsChanges: types.AllowedParamsChanges{tc.permission},
			}
			proposal := paramsproposal.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramsproposal.ParamChange{tc.paramChange},
			)
			s.Require().Equal(
				tc.expected,
				permission.Allows(s.ctx, s.pk, proposal),
			)
		})
	}
}

func (s *ParamsChangeTestSuite) TestAllowedParamsChange_InvalidJSON() {
	subspace, found := s.pk.GetSubspace(cdptypes.ModuleName)
	s.Require().True(found)
	subspace.Set(s.ctx, cdptypes.KeyDebtParam, s.cdpDebtParam)

	permission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{{
			Subspace:                   cdptypes.ModuleName,
			Key:                        string(cdptypes.KeyDebtParam),
			SingleSubparamAllowedAttrs: []string{"denom", "reference_asset", "conversion_factor", "debt_floor"},
		}},
	}
	proposal := paramsproposal.NewParameterChangeProposal(
		"A Title",
		"A description of this proposal.",
		[]paramsproposal.ParamChange{
			{
				Subspace: "cdp",
				Key:      "DebtParam",
				Value:    `{badjson}`,
			},
		},
	)
	s.Require().Equal(
		false,
		permission.Allows(s.ctx, s.pk, proposal),
	)
}

func (s *ParamsChangeTestSuite) TestAllowedParamsChange_InvalidJSONArray() {
	subspace, found := s.pk.GetSubspace(cdptypes.ModuleName)
	s.Require().True(found)
	subspace.Set(s.ctx, cdptypes.KeyCollateralParams, s.cdpCollateralParams)
	permission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{{
			Subspace:                   cdptypes.ModuleName,
			Key:                        string(cdptypes.KeyCollateralParams),
			MultiSubparamsRequirements: s.cdpCollateralRequirements,
		}},
	}
	proposal := paramsproposal.NewParameterChangeProposal(
		"A Title",
		"A description of this proposal.",
		[]paramsproposal.ParamChange{
			{
				Subspace: "cdp",
				Key:      string(cdptypes.KeyCollateralParams),
				Value:    `[badjson]`,
			},
		},
	)
	s.Require().Equal(
		false,
		permission.Allows(s.ctx, s.pk, proposal),
	)
}

func (s *ParamsChangeTestSuite) TestAllowedParamsChange_NoSubspaceData() {
	permission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{{
			Subspace:                   cdptypes.ModuleName,
			Key:                        string(cdptypes.KeyDebtParam),
			SingleSubparamAllowedAttrs: []string{"denom"},
		}},
	}
	proposal := paramsproposal.NewParameterChangeProposal(
		"A Title",
		"A description of this proposal.",
		[]paramsproposal.ParamChange{{
			Subspace: cdptypes.ModuleName,
			Key:      string(cdptypes.KeyDebtParam),
			Value:    `{}`,
		}},
	)
	s.Require().Panics(func() {
		permission.Allows(s.ctx, s.pk, proposal)
	})
}

func (s *ParamsChangeTestSuite) TestParamsChangePermission_NoAllowedChanged() {
	permission := types.ParamsChangePermission{}
	proposal := paramsproposal.NewParameterChangeProposal(
		"A Title",
		"A description of this proposal.",
		[]paramsproposal.ParamChange{
			{
				Key:      string(cdptypes.KeyDebtParam),
				Subspace: cdptypes.ModuleName,
				Value:    `{}`,
			},
		},
	)
	s.Require().False(permission.Allows(s.ctx, s.pk, proposal))
}

func (s *ParamsChangeTestSuite) TestParamsChangePermission_PassWhenOneAllowed() {
	subspace, found := s.pk.GetSubspace(cdptypes.ModuleName)
	s.Require().True(found)
	subspace.Set(s.ctx, cdptypes.KeyDebtParam, s.cdpDebtParam)

	permission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{
			{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"denom"},
			},
			{
				Subspace:                   cdptypes.ModuleName,
				Key:                        string(cdptypes.KeyDebtParam),
				SingleSubparamAllowedAttrs: []string{"reference_asset"},
			},
		},
	}
	proposal := paramsproposal.NewParameterChangeProposal(
		"A Title",
		"A description of this proposal.",
		// test success if one AllowedParamsChange is allowed and the other is not
		[]paramsproposal.ParamChange{
			{
				Key:      string(cdptypes.KeyDebtParam),
				Subspace: cdptypes.ModuleName,
				Value: `{
					"denom": "usdx",
					"reference_asset": "usd2",
					"conversion_factor": "6",
					"debt_floor": "1000"
				}`,
			},
		},
	)
	s.Require().True(permission.Allows(s.ctx, s.pk, proposal))
}

// Test subparam value with slice data unchanged comparision
func (s *ParamsChangeTestSuite) TestParamsChangePermission_SliceSubparamComparision() {
	permission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{{
			Subspace: pricefeedtypes.ModuleName,
			Key:      string(pricefeedtypes.KeyMarkets),
			MultiSubparamsRequirements: []types.SubparamRequirement{
				{
					Key:                        "market_id",
					Val:                        "xrp:usd",
					AllowedSubparamAttrChanges: []string{"quote_asset", "oracles"},
				},
				{
					Key:                        "market_id",
					Val:                        "btc:usd",
					AllowedSubparamAttrChanges: []string{"active"},
				},
			},
		}},
	}
	_, oracles := app.GeneratePrivKeyAddressPairs(5)

	testcases := []struct {
		name     string
		expected bool
		value    string
	}{
		{
			name:     "success changing allowed attrs",
			expected: true,
			value: fmt.Sprintf(`[{
				"market_id": "xrp:usd",
				"base_asset": "xrp",
				"quote_asset": "usdx",
				"oracles": [],
				"active": true
			},
			{
				"market_id": "btc:usd",
				"base_asset": "btc",
				"quote_asset": "usd",
				"oracles": ["%s"],
				"active": false
			}]`, oracles[1].String()),
		},
		{
			name:     "fails when changing not allowed attr (oracles)",
			expected: false,
			value: fmt.Sprintf(`[{
				"market_id": "xrp:usd",
				"base_asset": "xrp",
				"quote_asset": "usdx",
				"oracles": ["%s"],
				"active": true
			},
			{
				"market_id": "btc:usd",
				"base_asset": "btc",
				"quote_asset": "usd",
				"oracles": ["%s"],
				"active": false
			}]`, oracles[0].String(), oracles[2].String()),
		},
	}
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			subspace, found := s.pk.GetSubspace(pricefeedtypes.ModuleName)
			s.Require().True(found)
			currentMs := pricefeedtypes.Markets{
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{oracles[0]}, Active: true},
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{oracles[1]}, Active: true},
			}
			subspace.Set(s.ctx, pricefeedtypes.KeyMarkets, &currentMs)

			proposal := paramsproposal.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramsproposal.ParamChange{{
					Subspace: pricefeedtypes.ModuleName,
					Key:      string(pricefeedtypes.KeyMarkets),
					Value:    tc.value,
				}},
			)
			s.Require().Equal(
				tc.expected,
				permission.Allows(s.ctx, s.pk, proposal),
			)
		})
	}
}

func (s *ParamsChangeTestSuite) TestParamsChangePermission_NoSubparamRequirements() {
	permission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{{
			Subspace: cdptypes.ModuleName,
			Key:      string(cdptypes.KeySurplusThreshold),
		}},
	}

	testcases := []struct {
		name     string
		expected bool
		changes  []paramsproposal.ParamChange
	}{
		{
			name:     "success when changing allowed params",
			expected: true,
			changes: []paramsproposal.ParamChange{{
				Subspace: cdptypes.ModuleName,
				Key:      string(cdptypes.KeySurplusThreshold),
				Value:    sdkmath.NewInt(120).String(),
			}},
		},
		{
			name:     "fail when changing not allowed params",
			expected: false,
			changes: []paramsproposal.ParamChange{{
				Subspace: cdptypes.ModuleName,
				Key:      string(cdptypes.KeySurplusLot),
				Value:    sdkmath.NewInt(120).String(),
			}},
		},
		{
			name:     "fail if one change is not allowed",
			expected: false,
			changes: []paramsproposal.ParamChange{
				{
					Subspace: cdptypes.ModuleName,
					Key:      string(cdptypes.KeySurplusThreshold),
					Value:    sdkmath.NewInt(120).String(),
				},
				{
					Subspace: cdptypes.ModuleName,
					Key:      string(cdptypes.KeySurplusLot),
					Value:    sdkmath.NewInt(120).String(),
				},
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			subspace, found := s.pk.GetSubspace(cdptypes.ModuleName)
			s.Require().True(found)
			subspace.Set(s.ctx, cdptypes.KeySurplusThreshold, sdkmath.NewInt(100))
			subspace.Set(s.ctx, cdptypes.KeySurplusLot, sdkmath.NewInt(110))

			proposal := paramsproposal.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				tc.changes,
			)
			s.Require().Equal(
				tc.expected,
				permission.Allows(s.ctx, s.pk, proposal),
			)
		})
	}
}

func TestParamsChangeTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsChangeTestSuite))
}

func TestAllowedParamsChanges_Get(t *testing.T) {
	exampleAPCs := types.AllowedParamsChanges{
		{
			Subspace:                   "subspaceA",
			Key:                        "key1",
			SingleSubparamAllowedAttrs: []string{"attribute1"},
		},
		{
			Subspace:                   "subspaceA",
			Key:                        "key2",
			SingleSubparamAllowedAttrs: []string{"attribute2"},
		},
	}

	type args struct {
		subspace, key string
	}
	testCases := []struct {
		name  string
		apcs  types.AllowedParamsChanges
		args  args
		found bool
		out   types.AllowedParamsChange
	}{
		{
			name: "when element exists it is found",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceA",
				key:      "key2",
			},
			found: true,
			out:   exampleAPCs[1],
		},
		{
			name: "when element doesn't exist it isn't found",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceB",
				key:      "key1",
			},
			found: false,
		},
		{
			name: "when slice is nil, no elements are found",
			apcs: nil,
			args: args{
				subspace: "",
				key:      "",
			},
			found: false,
		},
		{
			name: "when slice is empty, no elements are found",
			apcs: types.AllowedParamsChanges{},
			args: args{
				subspace: "subspaceA",
				key:      "key1",
			},
			found: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, found := tc.apcs.Get(tc.args.subspace, tc.args.key)
			require.Equal(t, tc.found, found)
			require.Equal(t, tc.out, out)
		})
	}
}

func TestAllowedParamsChanges_Set(t *testing.T) {
	exampleAPCs := types.AllowedParamsChanges{
		{
			Subspace:                   "subspaceA",
			Key:                        "key1",
			SingleSubparamAllowedAttrs: []string{"attribute1"},
		},
		{
			Subspace:                   "subspaceA",
			Key:                        "key2",
			SingleSubparamAllowedAttrs: []string{"attribute2"},
		},
	}

	type args struct {
		subspace, key string
	}
	testCases := []struct {
		name string
		apcs types.AllowedParamsChanges
		arg  types.AllowedParamsChange
		out  types.AllowedParamsChanges
	}{
		{
			name: "when element isn't present it is added",
			apcs: exampleAPCs,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceB",
				Key:                        "key1",
				SingleSubparamAllowedAttrs: []string{"attribute1"},
			},
			out: append(exampleAPCs, types.AllowedParamsChange{
				Subspace:                   "subspaceB",
				Key:                        "key1",
				SingleSubparamAllowedAttrs: []string{"attribute1"},
			}),
		},
		{
			name: "when element matches, it is overwritten",
			apcs: exampleAPCs,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key1",
					SingleSubparamAllowedAttrs: []string{"attribute1"},
				},
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
		{
			name: "when element matches, it is overwritten",
			apcs: exampleAPCs,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key1",
					SingleSubparamAllowedAttrs: []string{"attribute1"},
				},
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
		{
			name: "when slice is nil, elements are added",
			apcs: nil,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
		{
			name: "when slice is empty, elements are added",
			apcs: types.AllowedParamsChanges{},
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			(&tc.apcs).Set(tc.arg)
			require.Equal(t, tc.out, tc.apcs)
		})
	}
}

func TestAllowedParamsChanges_Delete(t *testing.T) {
	exampleAPCs := types.AllowedParamsChanges{
		{
			Subspace:                   "subspaceA",
			Key:                        "key1",
			SingleSubparamAllowedAttrs: []string{"attribute1"},
		},
		{
			Subspace:                   "subspaceA",
			Key:                        "key2",
			SingleSubparamAllowedAttrs: []string{"attribute2"},
		},
	}

	type args struct {
		subspace, key string
	}
	testCases := []struct {
		name string
		apcs types.AllowedParamsChanges
		args args
		out  types.AllowedParamsChanges
	}{
		{
			name: "when element exists it is removed",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceA",
				key:      "key2",
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key1",
					SingleSubparamAllowedAttrs: []string{"attribute1"},
				},
			},
		},
		{
			name: "when element doesn't exist, none are removed",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceB",
				key:      "key1",
			},
			out: exampleAPCs,
		},
		{
			name: "when slice is nil, nothing happens",
			apcs: nil,
			args: args{
				subspace: "subspaceA",
				key:      "key1",
			},
			out: nil,
		},
		{
			name: "when slice is empty, nothing happens",
			apcs: types.AllowedParamsChanges{},
			args: args{
				subspace: "subspaceA",
				key:      "key1",
			},
			out: types.AllowedParamsChanges{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			(&tc.apcs).Delete(tc.args.subspace, tc.args.key)
			require.Equal(t, tc.out, tc.apcs)
		})
	}
}
