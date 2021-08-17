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
					DefaultRewardPeriods,
					Multipliers{
						NewMultiplier(Small, 1, sdk.MustNewDecFromStr("0.33")),
					},
					time.Date(2025, 10, 15, 14, 0, 0, 0, time.UTC),
				),
				USDXAccumulationTimes: GenesisAccumulationTimes{{
					CollateralType:           "bnb-a",
					PreviousAccumulationTime: time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
				}},
				USDXRewardIndexes: GenesisRewardIndexesSlice{{
					CollateralType: "bnb-a",
					RewardIndexes:  normalRewardIndexes[:1],
				}},
				USDXMintingClaims: USDXMintingClaims{
					{
						BaseClaim: BaseClaim{
							Owner:  sdk.AccAddress(crypto.AddressHash([]byte("KavaTestUser1"))),
							Reward: sdk.NewCoin("ukava", sdk.NewInt(100000000)),
						},
						RewardIndexes: []RewardIndex{
							{
								CollateralType: "bnb-a",
								RewardFactor:   sdk.ZeroDec(),
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
				USDXAccumulationTimes: GenesisAccumulationTimes{
					{
						CollateralType:           "",
						PreviousAccumulationTime: time.Date(2020, 10, 15, 14, 0, 0, 0, time.UTC),
					},
				},
				USDXMintingClaims: DefaultUSDXClaims,
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "collateral type must be defined",
			},
		},
		{
			name: "invalid claim",
			genesis: GenesisState{
				Params:                DefaultParams(),
				USDXAccumulationTimes: DefaultGenesisAccumulationTimes,
				USDXMintingClaims: USDXMintingClaims{
					{
						BaseClaim: BaseClaim{
							Owner:  sdk.AccAddress{},
							Reward: sdk.NewCoin("ukava", sdk.NewInt(100000000)),
						},
						RewardIndexes: []RewardIndex{
							{
								CollateralType: "bnb-a",
								RewardFactor:   sdk.ZeroDec(),
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
			name: "too many USDX reward factors is invalid",
			genesis: GenesisState{
				USDXRewardIndexes: GenesisRewardIndexesSlice{{
					CollateralType: "bnb-a",
					RewardIndexes: RewardIndexes{
						{
							CollateralType: "ukava",
							RewardFactor:   sdk.ZeroDec(),
						},
						{
							CollateralType: "hard",
							RewardFactor:   sdk.ZeroDec(),
						},
					},
				}},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "USDX reward indexes cannot have more than one reward denom",
			},
		},
		{
			name: "too many Delegator reward factors is invalid",
			genesis: GenesisState{
				HardDelegatorRewardIndexes: GenesisRewardIndexesSlice{{
					CollateralType: "ukava",
					RewardIndexes: RewardIndexes{
						{
							CollateralType: "ukava",
							RewardFactor:   sdk.ZeroDec(),
						},
						{
							CollateralType: "hard",
							RewardFactor:   sdk.ZeroDec(),
						},
					},
				}},
			},
			errArgs: errArgs{
				expectPass: false,
				contains:   "Delegator reward indexes cannot have more than one reward denom",
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

func TestGenesisRewardIndexeses_Validate(t *testing.T) {

	testCases := []struct {
		name    string
		indexes GenesisRewardIndexesSlice
		wantErr bool
	}{
		{
			name: "normal case",
			indexes: GenesisRewardIndexesSlice{
				{CollateralType: "btcb", RewardIndexes: normalRewardIndexes},
				{CollateralType: "bnb", RewardIndexes: normalRewardIndexes},
			},
			wantErr: false,
		},
		{
			name:    "empty",
			indexes: nil,
			wantErr: false,
		},
		{
			name:    "empty collateral type",
			indexes: GenesisRewardIndexesSlice{{RewardIndexes: normalRewardIndexes}},
			wantErr: true,
		},
		{
			name:    "invalid reward index",
			indexes: GenesisRewardIndexesSlice{{CollateralType: "btcb", RewardIndexes: invalidRewardIndexes}},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.indexes.Validate()
			if tc.wantErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
func TestGenesisAccumulationTimes_Validate(t *testing.T) {

	testCases := []struct {
		name    string
		gats    GenesisAccumulationTimes
		wantErr bool
	}{
		{
			name: "normal",
			gats: GenesisAccumulationTimes{
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
			gats: GenesisAccumulationTimes{
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

var normalRewardIndexes = RewardIndexes{
	NewRewardIndex("hard", sdk.MustNewDecFromStr("0.000001")),
	NewRewardIndex("ukava", sdk.MustNewDecFromStr("0.1")),
}

var invalidRewardIndexes = RewardIndexes{
	RewardIndex{"hard", sdk.MustNewDecFromStr("-0.01")},
}

var normalAccumulationtime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
