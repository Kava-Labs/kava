package v2_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	v2 "github.com/kava-labs/kava/x/incentive/migrations/v2"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
)

var kava10ParamsJson = `{
	"usdx_minting_reward_periods": null,
	"hard_supply_reward_periods": [
		{
			"active": true,
			"collateral_type": "usdx",
			"start": "2022-04-29T00:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "393201"
				},
				{
					"denom": "ukava",
					"amount": "491574"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			"start": "2022-04-29T00:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "88787"
				},
				{
					"denom": "ukava",
					"amount": "44688"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			"start": "2022-02-26T02:00:00Z",
			"end": "2022-05-26T02:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "38052"
				},
				{
					"denom": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
					"amount": "38052"
				},
				{
					"denom": "ukava",
					"amount": "3799"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "hard",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "31710"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "bnb",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "20611"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "busd",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "6342"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "btcb",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "31710"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "xrpb",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "17440"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ukava",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "6342"
				}
			]
		}
	],
	"hard_borrow_reward_periods": null,
	"delegator_reward_periods": [
		{
			"active": true,
			"collateral_type": "ukava",
			"start": "2022-02-26T01:30:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "316880"
				},
				{
					"denom": "swp",
					"amount": "198186"
				}
			]
		}
	],
	"swap_reward_periods": [
		{
			"active": true,
			"collateral_type": "bnb:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "btcb:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "busd:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "136986"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "hard:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B:usdx",
			"start": "2022-03-31T00:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "15221"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2:usdx",
			"start": "2022-03-31T00:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "53272"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098:usdx",
			"start": "2022-03-31T00:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "7610"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "swp:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "152207"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ukava:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "167428"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "usdx:xrpb",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		}
	],
	"claim_multipliers": [
		{
			"denom": "hard",
			"multipliers": [
				{
					"name": "small",
					"months_lockup": "1",
					"factor": "0.200000000000000000"
				},
				{
					"name": "large",
					"months_lockup": "12",
					"factor": "1.000000000000000000"
				}
			]
		},
		{
			"denom": "ukava",
			"multipliers": [
				{
					"name": "small",
					"months_lockup": "1",
					"factor": "0.200000000000000000"
				},
				{
					"name": "large",
					"months_lockup": "12",
					"factor": "1.000000000000000000"
				}
			]
		},
		{
			"denom": "swp",
			"multipliers": [
				{
					"name": "small",
					"months_lockup": "1",
					"factor": "0.100000000000000000"
				},
				{
					"name": "large",
					"months_lockup": "12",
					"factor": "1.000000000000000000"
				}
			]
		},
		{
			"denom": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			"multipliers": [
				{
					"name": "large",
					"factor": "1.000000000000000000"
				}
			]
		}
	],
	"claim_end": "2026-04-08T14:00:00Z",
	"savings_reward_periods": null
}`

var kava11ParamsJson = `{
	"usdx_minting_reward_periods": [],
	"hard_supply_reward_periods": [
		{
			"active": true,
			"collateral_type": "usdx",
			"start": "2022-04-29T00:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "393201"
				},
				{
					"denom": "ukava",
					"amount": "473532"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
			"start": "2022-04-29T00:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "88787"
				},
				{
					"denom": "ukava",
					"amount": "11098"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			"start": "2022-02-26T02:00:00Z",
			"end": "2022-05-26T02:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "38052"
				},
				{
					"denom": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
					"amount": "38052"
				},
				{
					"denom": "ukava",
					"amount": "3799"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "hard",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "31710"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "bnb",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "20611"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "busd",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "6342"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "btcb",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "31710"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "xrpb",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "17440"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ukava",
			"start": "2020-10-16T14:00:00Z",
			"end": "2024-10-16T14:00:00Z",
			"rewards_per_second": [
				{
					"denom": "hard",
					"amount": "6342"
				}
			]
		}
	],
	"hard_borrow_reward_periods": [],
	"delegator_reward_periods": [],
	"swap_reward_periods": [
		{
			"active": true,
			"collateral_type": "bnb:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "btcb:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "busd:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "136986"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "hard:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/0471F1C4E7AFD3F07702BEF6DC365268D64570F7C1FDC98EA6098DD6DE59817B:usdx",
			"start": "2022-03-31T00:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "15221"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2:usdx",
			"start": "2022-03-31T00:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "53272"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098:usdx",
			"start": "2022-03-31T00:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "7610"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "swp:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "152207"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "ukava:usdx",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "167428"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "usdx:xrpb",
			"start": "2021-08-30T15:00:00Z",
			"end": "2025-08-29T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "swp",
					"amount": "38052"
				}
			]
		}
	],
	"claim_multipliers": [
		{
			"denom": "hard",
			"multipliers": [
				{
					"name": "large",
					"months_lockup": "0",
					"factor": "1.000000000000000000"
				}
			]
		},
		{
			"denom": "ukava",
			"multipliers": [
				{
					"name": "large",
					"months_lockup": "0",
					"factor": "1.000000000000000000"
				}
			]
		},
		{
			"denom": "swp",
			"multipliers": [
				{
					"name": "large",
					"months_lockup": "0",
					"factor": "1.000000000000000000"
				}
			]
		},
		{
			"denom": "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			"multipliers": [
				{
					"name": "large",
					"months_lockup": "0",
					"factor": "1.000000000000000000"
				}
			]
		}
	],
	"earn_reward_periods": [
		{
			"active": true,
			"collateral_type": "bkava",
			"start": "2022-10-26T15:00:00Z",
			"end": "2024-10-26T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "ukava",
					"amount": "190258"
				}
			]
		},
		{
			"active": true,
			"collateral_type": "erc20/multichain/usdc",
			"start": "2022-10-26T15:00:00Z",
			"end": "2024-10-26T15:00:00Z",
			"rewards_per_second": [
				{
					"denom": "ukava",
					"amount": "5284"
				}
			]
		}
	],
	"claim_end": "2026-04-08T14:00:00Z",
	"savings_reward_periods": []
}`

type MigrateTestSuite struct {
	testutil.IntegrationTester

	genesisTime time.Time
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateTestSuite))
}

func (suite *MigrateTestSuite) SetupTest() {
	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *MigrateTestSuite) TestMigrate() {
	suite.SetApp()
	cdc := suite.App.AppCodec()

	var kava10Params types.Params
	cdc.MustUnmarshalJSON([]byte(kava10ParamsJson), &kava10Params)

	var kava11Params types.Params
	cdc.MustUnmarshalJSON([]byte(kava11ParamsJson), &kava11Params)

	suite.Require().NoError(kava10Params.Validate())
	suite.Require().NoError(kava11Params.Validate())

	incentiveGs := types.NewGenesisState(
		kava10Params,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultGenesisRewardState,
		types.DefaultUSDXClaims,
		types.DefaultHardClaims,
		types.DefaultDelegatorClaims,
		types.DefaultSwapClaims,
		types.DefaultSavingsClaims,
		types.DefaultEarnClaims,
	)

	suite.StartChain(
		app.GenesisState{
			types.ModuleName: cdc.MustMarshalJSON(&incentiveGs),
		},
	)

	subspace, found := suite.App.GetParamsKeeper().GetSubspace(types.ModuleName)
	suite.Require().True(found)

	err := v2.MigrateParams(suite.Ctx, subspace)
	suite.Require().NoError(err)

	newParams := suite.App.GetIncentiveKeeper().GetParams(suite.Ctx)

	suite.Require().JSONEq(
		kava11ParamsJson,
		string(cdc.MustMarshalJSON(&newParams)),
		"migrated params should match expected params",
	)

	// Extra checks for expected changes

	suite.Nil(newParams.DelegatorRewardPeriods, "delegator reward periods should be removed")

	for _, multiplier := range newParams.ClaimMultipliers {
		suite.Len(multiplier.Multipliers, 1)
		suite.Equal(
			types.Multiplier{
				Name:         "large",
				MonthsLockup: 0,
				Factor:       sdk.OneDec(),
			},
			multiplier.Multipliers[0],
			"all multipliers should be \"large\" with 0 months lockup and factor 1",
		)
	}

	suite.Len(newParams.EarnRewardPeriods, 2)
}
