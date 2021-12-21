package types_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	types "github.com/kava-labs/kava/x/committee/types"
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
		ConversionFactor: sdk.NewInt(6),
		DebtFloor:        sdk.NewInt(1000),
	}

	suite.cdpCollateralParams = cdptypes.CollateralParams{
		{
			Denom:                            "bnb",
			Type:                             "bnb-a",
			LiquidationRatio:                 sdk.MustNewDecFromStr("2.0"),
			DebtLimit:                        sdk.NewCoin("usdx", sdk.NewInt(100)),
			StabilityFee:                     sdk.MustNewDecFromStr("1.02"),
			LiquidationPenalty:               sdk.MustNewDecFromStr("0.05"),
			AuctionSize:                      sdk.NewInt(100),
			ConversionFactor:                 sdk.NewInt(6),
			SpotMarketID:                     "bnb:usd",
			LiquidationMarketID:              "bnb:usd",
			CheckCollateralizationIndexCount: sdk.NewInt(0),
		},
		{
			Denom:                            "btc",
			Type:                             "btc-a",
			LiquidationRatio:                 sdk.MustNewDecFromStr("1.5"),
			DebtLimit:                        sdk.NewCoin("usdx", sdk.NewInt(100)),
			StabilityFee:                     sdk.MustNewDecFromStr("1.01"),
			LiquidationPenalty:               sdk.MustNewDecFromStr("0.10"),
			AuctionSize:                      sdk.NewInt(1000),
			ConversionFactor:                 sdk.NewInt(8),
			SpotMarketID:                     "btc:usd",
			LiquidationMarketID:              "btc:usd",
			CheckCollateralizationIndexCount: sdk.NewInt(1),
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

// Test fail when param change value is invalid json
// Test fail when param change value is invalid array json
// Test fail if bad existing data in subspace
// Test success if permissions does not contain any changes
// succeeds if one AllowedParamsChange is allowed and the other is not
// succeeds if changing non cdp subspace data
// fails if proposal has no corresponding allowed param changes in the permission
func TestParamsChangeTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsChangeTestSuite))
}
