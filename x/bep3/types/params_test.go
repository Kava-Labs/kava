package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/bep3/types"
)

type ParamsTestSuite struct {
	suite.Suite
	addr sdk.AccAddress
}

func (suite *ParamsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	suite.addr = addrs[0]
	return
}

func (suite *ParamsTestSuite) TestParamValidation() {
	type LoadParams func() types.Params

	type args struct {
		bnbDeputyAddress  sdk.AccAddress
		bnbDeputyFixedFee uint64
		minAmount         uint64
		maxAmount         uint64
		minBlockLock      uint64
		maxBlockLock      uint64
		supportedAssets   types.AssetParams
	}

	testCases := []struct {
		name        string
		args        args
		expectPass  bool
		expectedErr string
	}{
		{
			name: "default",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      types.DefaultMinBlockLock,
				maxBlockLock:      types.DefaultMaxBlockLock,
				supportedAssets:   types.DefaultSupportedAssets,
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "minimum block lock == maximum block lock",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      243,
				maxBlockLock:      243,
				supportedAssets:   types.DefaultSupportedAssets,
			},
			expectPass:  true,
			expectedErr: "",
		},
		{
			name: "minimum amount greater than maximum amount",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         10000000,
				maxAmount:         100000,
				minBlockLock:      types.DefaultMinBlockLock,
				maxBlockLock:      types.DefaultMaxBlockLock,
				supportedAssets:   types.DefaultSupportedAssets,
			},
			expectPass:  false,
			expectedErr: "minimum amount cannot be > maximum amount",
		},
		{
			name: "minimum block lock greater than maximum block lock",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      500,
				maxBlockLock:      400,
				supportedAssets:   types.DefaultSupportedAssets,
			},
			expectPass:  false,
			expectedErr: "minimum block lock cannot be > maximum block lock",
		},
		{
			name: "empty asset denom",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      types.DefaultMinBlockLock,
				maxBlockLock:      types.DefaultMaxBlockLock,
				supportedAssets: types.AssetParams{
					types.AssetParam{
						Denom:  "",
						CoinID: 714,
						Limit:  sdk.NewInt(100000000000),
						Active: true,
					},
				},
			},
			expectPass:  false,
			expectedErr: "asset denom cannot be empty",
		},
		{
			name: "negative asset coin ID",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      types.DefaultMinBlockLock,
				maxBlockLock:      types.DefaultMaxBlockLock,
				supportedAssets: types.AssetParams{
					types.AssetParam{
						Denom:  "bnb",
						CoinID: -1,
						Limit:  sdk.NewInt(100000000000),
						Active: true,
					},
				},
			},
			expectPass:  false,
			expectedErr: "must be a non negative integer",
		},
		{
			name: "negative asset limit",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      types.DefaultMinBlockLock,
				maxBlockLock:      types.DefaultMaxBlockLock,
				supportedAssets: types.AssetParams{
					types.AssetParam{
						Denom:  "bnb",
						CoinID: 714,
						Limit:  sdk.NewInt(-10000),
						Active: true,
					},
				},
			},
			expectPass:  false,
			expectedErr: "must have a positive supply limit",
		},
		{
			name: "duplicate asset denom",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      types.DefaultMinBlockLock,
				maxBlockLock:      types.DefaultMaxBlockLock,
				supportedAssets: types.AssetParams{
					types.AssetParam{
						Denom:  "bnb",
						CoinID: 714,
						Limit:  sdk.NewInt(100000000000),
						Active: true,
					},
					types.AssetParam{
						Denom:  "bnb",
						CoinID: 114,
						Limit:  sdk.NewInt(500000000),
						Active: false,
					},
				},
			},
			expectPass:  false,
			expectedErr: "cannot have duplicate denom",
		},
		{
			name: "duplicate asset coin ID",
			args: args{
				bnbDeputyAddress:  suite.addr,
				bnbDeputyFixedFee: types.DefaultBnbDeputyFixedFee,
				minAmount:         types.DefaultMinAmount,
				maxAmount:         types.DefaultMaxAmount,
				minBlockLock:      types.DefaultMinBlockLock,
				maxBlockLock:      types.DefaultMaxBlockLock,
				supportedAssets: types.AssetParams{
					types.AssetParam{
						Denom:  "bnb",
						CoinID: 714,
						Limit:  sdk.NewInt(100000000000),
						Active: true,
					},
					types.AssetParam{
						Denom:  "fake",
						CoinID: 714,
						Limit:  sdk.NewInt(500000000),
						Active: false,
					},
				},
			},
			expectPass:  false,
			expectedErr: "cannot have duplicate coin id",
		},
	}

	for _, tc := range testCases {
		params := types.NewParams(tc.args.bnbDeputyAddress, tc.args.bnbDeputyFixedFee, tc.args.minAmount,
			tc.args.maxAmount, tc.args.minBlockLock, tc.args.maxBlockLock, tc.args.supportedAssets)

		err := params.Validate()
		if tc.expectPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
			suite.Require().Contains(err.Error(), tc.expectedErr)
		}
	}
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
