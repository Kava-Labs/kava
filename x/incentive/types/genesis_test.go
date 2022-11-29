package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestGenesisState_Validate(t *testing.T) {
	type errArgs struct {
		expectPass bool
		contains   string
	}

	testCases := []struct {
		name    string
		genesis GenesisState
		errArgs errArgs
	}{
		{
			name:    "default",
			genesis: DefaultGenesisState(),
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "valid",
			genesis: GenesisState{
				Params: NewParams(
					RewardPeriods{
						NewRewardPeriod(
							true,
							"bnb-a",
							time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
							time.Date(2024, 10, 15, 14, 0, 0, 0, time.UTC),
							sdk.NewCoin("ukava", sdk.NewInt(25000)),
						),
					},
					DefaultMultiRewardPeriods,
					DefaultMultiRewardPeriods,
					DefaultMultiRewardPeriods,
					DefaultMultiRewardPeriods,
					DefaultMultiRewardPeriods,
					DefaultMultiRewardPeriods,
					MultipliersPerDenoms{
						{
							Denom: "ukava",
							Multipliers: Multipliers{
								NewMultiplier("small", 1, sdk.MustNewDecFromStr("0.33")),
								NewMultiplier("large", 12, sdk.MustNewDecFromStr("1.00")),
							},
						},
					},
					time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
					DefaultTypedMultiRewardPeriods,
				),
				USDXRewardState: GenesisRewardState{
					AccumulationTimes: AccumulationTimes{{
						CollateralType:           "bnb-a",
						PreviousAccumulationTime: time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
					}},
					MultiRewardIndexes: MultiRewardIndexes{{
						CollateralType: "bnb-a",
						RewardIndexes:  normalRewardIndexes,
					}},
				},
				Claims: Claims{
					{
						Type:   CLAIM_TYPE_USDX_MINTING,
						Owner:  sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser1"))),
						Reward: cs(c("ukava", 100000000)),
						RewardIndexes: []MultiRewardIndex{
							{
								CollateralType: "bnb-a",
								RewardIndexes: RewardIndexes{
									NewRewardIndex(USDXMintingRewardDenom, sdk.ZeroDec()),
								},
							},
						},
					},
				},
			},
			errArgs: errArgs{
				expectPass: true,
			},
		},
		{
			name: "invalid genesis accumulation time",
			genesis: GenesisState{
				Params: DefaultParams(),
				USDXRewardState: GenesisRewardState{
					AccumulationTimes: AccumulationTimes{{
						CollateralType:           "",
						PreviousAccumulationTime: time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
					}},
					MultiRewardIndexes: MultiRewardIndexes{{
						CollateralType: "bnb-a",
						RewardIndexes:  normalRewardIndexes,
					}},
				},
				Claims: DefaultClaims,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "collateral type must be defined",
			},
		},
		{
			name: "invalid claim",
			genesis: GenesisState{
				Params:          DefaultParams(),
				USDXRewardState: DefaultGenesisRewardState,
				Claims: Claims{
					{
						Type:   CLAIM_TYPE_USDX_MINTING,
						Owner:  nil, // invalid address
						Reward: cs(c("ukava", 100000000)),
						RewardIndexes: []MultiRewardIndex{
							{
								CollateralType: "bnb-a",
								RewardIndexes: RewardIndexes{
									NewRewardIndex(USDXMintingRewardDenom, sdk.ZeroDec()),
								},
							},
						},
					},
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "claim owner cannot be empty",
			},
		},
		{
			name: "invalid accrual time",
			genesis: GenesisState{
				Params: DefaultParams(),
				Claims: DefaultClaims,
				AccrualTimes: AccrualTimes{
					NewAccrualTime(CLAIM_TYPE_USDX_MINTING, "", time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC)),
				},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "collateral type cannot be empty",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.errArgs.expectPass {
				require.NoError(t, err, tc.name)
			} else {
				require.Error(t, err, tc.name)
				require.Contains(t, err.Error(), tc.errArgs.contains)
			}
		})
	}
}

func TestGenesisAccumulationTimes_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		gats    AccumulationTimes
		wantErr bool
	}{
		{
			name: "normal",
			gats: AccumulationTimes{
				{CollateralType: "btcb", PreviousAccumulationTime: normalAccumulationtime},
				{CollateralType: "bnb", PreviousAccumulationTime: normalAccumulationtime},
			},
			wantErr: false,
		},
		{
			name:    "empty",
			gats:    nil,
			wantErr: false,
		},
		{
			name: "empty collateral type",
			gats: AccumulationTimes{
				{PreviousAccumulationTime: normalAccumulationtime},
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gats.Validate()
			if tc.wantErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestAccrualTimes_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		gats    AccrualTimes
		wantErr bool
	}{
		{
			name: "normal",
			gats: AccrualTimes{
				{
					ClaimType:                CLAIM_TYPE_USDX_MINTING,
					CollateralType:           "btcb",
					PreviousAccumulationTime: normalAccumulationtime,
				},
				{
					ClaimType:                CLAIM_TYPE_USDX_MINTING,
					CollateralType:           "bnb",
					PreviousAccumulationTime: normalAccumulationtime,
				},
			},
			wantErr: false,
		},
		{
			name:    "empty",
			gats:    nil,
			wantErr: false,
		},
		{
			name: "empty collateral type",
			gats: AccrualTimes{
				{
					ClaimType:                CLAIM_TYPE_USDX_MINTING,
					PreviousAccumulationTime: normalAccumulationtime,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid claim type",
			gats: AccrualTimes{
				{
					ClaimType:                10000000,
					CollateralType:           "btcb",
					PreviousAccumulationTime: normalAccumulationtime,
				},
			},
			wantErr: true,
		},
		{
			name: "unspecified claim type",
			gats: AccrualTimes{
				{
					ClaimType:                CLAIM_TYPE_UNSPECIFIED,
					CollateralType:           "btcb",
					PreviousAccumulationTime: normalAccumulationtime,
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gats.Validate()
			if tc.wantErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

var normalAccumulationtime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
