package e2e_test

import (
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (suite *IntegrationTestSuite) TestUpgradeParams_SDK() {
	suite.SkipIfUpgradeDisabled()

	beforeUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight - 1)
	afterUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight)

	// Before params
	grpcClient := suite.Kava.Grpc
	govParamsBefore, err := grpcClient.Query.Gov.Params(beforeUpgradeCtx, &govtypes.QueryParamsRequest{
		ParamsType: govtypes.ParamDeposit,
	})
	suite.NoError(err)
	govParamsAfter, err := grpcClient.Query.Gov.Params(afterUpgradeCtx, &govtypes.QueryParamsRequest{
		ParamsType: govtypes.ParamDeposit,
	})
	suite.NoError(err)

	// after upgrade, querying params before upgrade height returns nil
	// since the param gprc query no longer queries x/params
	suite.Run("x/gov parameters before upgrade", func() {
		suite.Assert().Nil(
			govParamsBefore.DepositParams.MaxDepositPeriod,
			"x/gov DepositParams max deposit period before upgrade should be nil",
		)
		suite.Assert().Nil(
			govParamsBefore.DepositParams.MinDeposit,
			"x/gov DepositParams min deposit before upgrade should be 10_000_000 ukava",
		)
	})

	suite.Run("x/gov parameters after upgrade", func() {
		suite.Assert().Equal(
			mustParseDuration("172800s"),
			govParamsAfter.DepositParams.MaxDepositPeriod,
			"x/gov DepositParams max deposit period after upgrade should be 172800s",
		)
		suite.Assert().Equal(
			[]sdk.Coin{{Denom: "ukava", Amount: sdk.NewInt(10_000_000)}},
			govParamsAfter.DepositParams.MinDeposit,
			"x/gov DepositParams min deposit after upgrade should be 10_000_000 ukava",
		)

		expectedParams := govtypes.Params{
			MinDeposit:                 sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(10_000_000))),
			MaxDepositPeriod:           mustParseDuration("172800s"),
			VotingPeriod:               mustParseDuration("30s"),
			Quorum:                     "0.334000000000000000",
			Threshold:                  "0.500000000000000000",
			VetoThreshold:              "0.334000000000000000",
			MinInitialDepositRatio:     "0.000000000000000000",
			BurnVoteQuorum:             false,
			BurnProposalDepositPrevote: false,
			BurnVoteVeto:               true,
		}
		suite.Require().Equal(expectedParams, *govParamsAfter.Params, "x/gov params after upgrade should be as expected")
	})
}

func (suite *IntegrationTestSuite) TestUpgradeParams_Consensus() {
	suite.SkipIfUpgradeDisabled()

	afterUpgradeCtx := suite.Kava.Grpc.CtxAtHeight(suite.UpgradeHeight)

	grpcClient := suite.Kava.Grpc
	paramsAfter, err := grpcClient.Query.Consensus.Params(afterUpgradeCtx, &consensustypes.QueryParamsRequest{})
	suite.NoError(err)

	// v25 consensus params from x/params should be migrated to x/consensus
	expectedParams := tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
			MaxBytes: 22020096,
			MaxGas:   20000000,
		},
		Evidence: &tmproto.EvidenceParams{
			MaxAgeNumBlocks: 100000,
			MaxAgeDuration:  *mustParseDuration("172800s"),
			MaxBytes:        1048576,
		},
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: []string{
				tmtypes.ABCIPubKeyTypeEd25519,
			},
		},
		Version: nil,
	}
	suite.Require().Equal(expectedParams, *paramsAfter.Params, "x/consensus params after upgrade should be as expected")
}

func mustParseDuration(s string) *time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return &d
}
