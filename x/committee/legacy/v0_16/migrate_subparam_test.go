package v0_16

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v016bep3types "github.com/kava-labs/kava/x/bep3/types"
	v016cdptypes "github.com/kava-labs/kava/x/cdp/types"
	v015committee "github.com/kava-labs/kava/x/committee/legacy/v0_15"
	v016committee "github.com/kava-labs/kava/x/committee/types"
	v016hardtypes "github.com/kava-labs/kava/x/hard/types"
	v016pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

func (s *migrateTestSuite) TestMigrate_Committee_SubparamPermissions() {
	testcases := []struct {
		name           string
		v015permission v015committee.Permission
		v016permission v016committee.Permission
	}{
		{
			name: "allowed collateral params",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016cdptypes.ModuleName,
					Key:      string(v016cdptypes.KeyCollateralParams),
				}},
				AllowedCollateralParams: v015committee.AllowedCollateralParams{
					{
						Type:                   "bnb",
						Denom:                  true,
						LiquidationRatio:       false,
						DebtLimit:              true,
						KeeperRewardPercentage: true,
					},
					{
						Type:                             "btc",
						Prefix:                           true,
						SpotMarketID:                     false,
						DebtLimit:                        true,
						CheckCollateralizationIndexCount: true,
					},
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace: v016cdptypes.ModuleName,
						Key:      string(v016cdptypes.KeyCollateralParams),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{
							{
								Key:                        "type",
								Val:                        "bnb",
								AllowedSubparamAttrChanges: []string{"debt_limit", "denom", "keeper_reward_percentage"},
							},
							{
								Key:                        "type",
								Val:                        "btc",
								AllowedSubparamAttrChanges: []string{"check_collateralization_index_count", "debt_limit", "prefix"},
							},
						},
					},
				},
			},
		},
		{
			name: "allowed collateral params - no requirements",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016cdptypes.ModuleName,
					Key:      string(v016cdptypes.KeyCollateralParams),
				}},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace:                   v016cdptypes.ModuleName,
						Key:                        string(v016cdptypes.KeyCollateralParams),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{},
					},
				},
			},
		},
		{
			name: "allowed debt params",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016cdptypes.ModuleName,
					Key:      string(v016cdptypes.KeyDebtParam),
				}},
				AllowedDebtParam: v015committee.AllowedDebtParam{
					Denom:            true,
					ReferenceAsset:   false,
					ConversionFactor: true,
					DebtFloor:        true,
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace:                   v016cdptypes.ModuleName,
						Key:                        string(v016cdptypes.KeyDebtParam),
						SingleSubparamAllowedAttrs: []string{"conversion_factor", "debt_floor", "denom"},
					},
				},
			},
		},
		{
			name: "allowed debt params - no requirements",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016cdptypes.ModuleName,
					Key:      string(v016cdptypes.KeyDebtParam),
				}},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace:                   v016cdptypes.ModuleName,
						Key:                        string(v016cdptypes.KeyDebtParam),
						SingleSubparamAllowedAttrs: []string{},
					},
				},
			},
		},
		{
			name: "param not allowed",
			v015permission: v015committee.SubParamChangePermission{
				AllowedDebtParam: v015committee.AllowedDebtParam{
					Denom:            true,
					ReferenceAsset:   false,
					ConversionFactor: true,
					DebtFloor:        true,
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{},
			},
		},
		{
			name: "allowed asset params",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016bep3types.ModuleName,
					Key:      string(v016bep3types.KeyAssetParams),
				}},
				AllowedAssetParams: v015committee.AllowedAssetParams{
					{
						Denom:         "bnb",
						CoinID:        true,
						MaxSwapAmount: true,
						Active:        true,
					},
					{
						Denom:        "btc",
						Limit:        true,
						MinBlockLock: true,
						Active:       true,
					},
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace: v016bep3types.ModuleName,
						Key:      string(v016bep3types.KeyAssetParams),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{
							{
								Key:                        "denom",
								Val:                        "bnb",
								AllowedSubparamAttrChanges: []string{"active", "coin_id", "max_swap_amount"},
							},
							{
								Key:                        "denom",
								Val:                        "btc",
								AllowedSubparamAttrChanges: []string{"active", "limit", "min_block_lock"},
							},
						},
					},
				},
			},
		},
		{
			name: "allowed asset params - no requirements",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016bep3types.ModuleName,
					Key:      string(v016bep3types.KeyAssetParams),
				}},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace:                   v016bep3types.ModuleName,
						Key:                        string(v016bep3types.KeyAssetParams),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{},
					},
				},
			},
		},
		{
			name: "allowed markets",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016pricefeedtypes.ModuleName,
					Key:      string(v016pricefeedtypes.KeyMarkets),
				}},
				AllowedMarkets: v015committee.AllowedMarkets{
					{
						MarketID:   "bnb-btc",
						BaseAsset:  false,
						QuoteAsset: true,
						Active:     true,
					},
					{
						MarketID:  "btc-usd",
						BaseAsset: true,
						Oracles:   true,
						Active:    true,
					},
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace: v016pricefeedtypes.ModuleName,
						Key:      string(v016pricefeedtypes.KeyMarkets),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{
							{
								Key:                        "market_id",
								Val:                        "bnb-btc",
								AllowedSubparamAttrChanges: []string{"active", "quote_asset"},
							},
							{
								Key:                        "market_id",
								Val:                        "btc-usd",
								AllowedSubparamAttrChanges: []string{"active", "base_asset", "oracles"},
							},
						},
					},
				},
			},
		},
		{
			name: "allowed markets - no requirements",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016pricefeedtypes.ModuleName,
					Key:      string(v016pricefeedtypes.KeyMarkets),
				}},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace:                   v016pricefeedtypes.ModuleName,
						Key:                        string(v016pricefeedtypes.KeyMarkets),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{},
					},
				},
			},
		},
		{
			name: "allowed money markets",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016hardtypes.ModuleName,
					Key:      string(v016hardtypes.KeyMoneyMarkets),
				}},
				AllowedMoneyMarkets: v015committee.AllowedMoneyMarkets{
					{
						Denom:                  "bnb",
						BorrowLimit:            true,
						ConversionFactor:       false,
						ReserveFactor:          true,
						KeeperRewardPercentage: true,
					},
					{
						Denom:             "btc",
						BorrowLimit:       false,
						SpotMarketID:      true,
						InterestRateModel: true,
					},
				},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace: v016hardtypes.ModuleName,
						Key:      string(v016hardtypes.KeyMoneyMarkets),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{
							{
								Key:                        "denom",
								Val:                        "bnb",
								AllowedSubparamAttrChanges: []string{"borrow_limit", "keeper_reward_percentage", "reserve_factor"},
							},
							{
								Key:                        "denom",
								Val:                        "btc",
								AllowedSubparamAttrChanges: []string{"interest_rate_model", "spot_market_id"},
							},
						},
					},
				},
			},
		},
		{
			name: "allowed money markets - no requirements",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016hardtypes.ModuleName,
					Key:      string(v016hardtypes.KeyMoneyMarkets),
				}},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace:                   v016hardtypes.ModuleName,
						Key:                        string(v016hardtypes.KeyMoneyMarkets),
						MultiSubparamsRequirements: []v016committee.SubparamRequirement{},
					},
				},
			},
		},
		{
			name: "allowed params",
			v015permission: v015committee.SubParamChangePermission{
				AllowedParams: v015committee.AllowedParams{{
					Subspace: v016hardtypes.ModuleName,
					Key:      string(v016hardtypes.KeyMinimumBorrowUSDValue),
				}},
			},
			v016permission: &v016committee.ParamsChangePermission{
				AllowedParamsChanges: v016committee.AllowedParamsChanges{
					{
						Subspace: v016hardtypes.ModuleName,
						Key:      string(v016hardtypes.KeyMinimumBorrowUSDValue),
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			oldCommittee := v015committee.MemberCommittee{
				BaseCommittee: v015committee.BaseCommittee{
					ID:               1,
					Description:      "test",
					Members:          s.addresses,
					Permissions:      []v015committee.Permission{tc.v015permission},
					VoteThreshold:    sdk.NewDec(40),
					ProposalDuration: time.Hour * 24 * 7,
					TallyOption:      v015committee.FirstPastThePost,
				},
			}
			expectedProposal, err := v016committee.NewMemberCommittee(1, "test", s.addresses, []v016committee.Permission{tc.v016permission}, oldCommittee.VoteThreshold, oldCommittee.ProposalDuration, v016committee.TALLY_OPTION_FIRST_PAST_THE_POST)
			s.Require().NoError(err)
			s.v15genstate.Committees = []v015committee.Committee{oldCommittee}
			genState := Migrate(s.v15genstate)
			s.Require().Len(genState.Committees, 1)
			s.Equal(expectedProposal, genState.GetCommittees()[0])
		})
	}
}
