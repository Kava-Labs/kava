package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/kava-labs/kava/app"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
	"github.com/kava-labs/kava/x/savings/keeper"
	"github.com/kava-labs/kava/x/savings/types"
)

var dep = types.NewDeposit

const (
	bkava1 = "bkava-kavavaloper15gqc744d05xacn4n6w2furuads9fu4pqn6zxlu"
	bkava2 = "bkava-kavavaloper15qdefkmwswysgg4qxgqpqr35k3m49pkx8yhpte"
)

type grpcQueryTestSuite struct {
	suite.Suite

	tApp        app.TestApp
	ctx         sdk.Context
	keeper      keeper.Keeper
	queryServer types.QueryServer
	addrs       []sdk.AccAddress
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.tApp = app.NewTestApp()
	_, addrs := app.GeneratePrivKeyAddressPairs(2)

	suite.addrs = addrs

	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	suite.keeper = suite.tApp.GetSavingsKeeper()
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)

	err := suite.tApp.FundModuleAccount(
		suite.ctx,
		types.ModuleAccountName,
		cs(
			c("usdx", 10000000000),
			c("busd", 10000000000),
		),
	)
	suite.Require().NoError(err)

	savingsGenesis := types.GenesisState{
		Params: types.NewParams([]string{"bnb", "busd", bkava1, bkava2}),
	}
	savingsGenState := app.GenesisState{types.ModuleName: suite.tApp.AppCodec().MustMarshalJSON(&savingsGenesis)}

	suite.tApp.InitializeFromGenesisStates(
		savingsGenState,
		app.NewFundedGenStateWithSameCoins(
			suite.tApp.AppCodec(),
			cs(
				c("bnb", 10000000000),
				c("busd", 20000000000),
			),
			addrs,
		),
	)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryParams() {
	res, err := suite.queryServer.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	var expected types.GenesisState
	savingsGenesis := types.GenesisState{
		Params: types.NewParams([]string{"bnb", "busd", bkava1, bkava2}),
	}
	savingsGenState := app.GenesisState{types.ModuleName: suite.tApp.AppCodec().MustMarshalJSON(&savingsGenesis)}
	suite.tApp.AppCodec().MustUnmarshalJSON(savingsGenState[types.ModuleName], &expected)

	suite.Equal(expected.Params, res.Params, "params should equal test genesis state")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryDeposits() {
	suite.addDeposits([]types.Deposit{
		dep(suite.addrs[0], cs(c("bnb", 100000000))),
		dep(suite.addrs[1], cs(c("bnb", 20000000))),
		dep(suite.addrs[0], cs(c("busd", 20000000))),
		dep(suite.addrs[0], cs(c("busd", 8000000))),
	})

	tests := []struct {
		giveName          string
		giveRequest       *types.QueryDepositsRequest
		wantDepositCounts int
		shouldError       bool
		errorSubstr       string
	}{
		{
			"empty query",
			&types.QueryDepositsRequest{},
			2,
			false,
			"",
		},
		{
			"owner",
			&types.QueryDepositsRequest{
				Owner: suite.addrs[0].String(),
			},
			// Excludes the second address
			1,
			false,
			"",
		},
		{
			"invalid owner",
			&types.QueryDepositsRequest{
				Owner: "invalid address",
			},
			// No deposits
			0,
			true,
			"decoding bech32 failed",
		},
		{
			"owner and denom",
			&types.QueryDepositsRequest{
				Owner: suite.addrs[0].String(),
				Denom: "bnb",
			},
			// Only the first one
			1,
			false,
			"",
		},
		{
			"owner and invalid denom empty response",
			&types.QueryDepositsRequest{
				Owner: suite.addrs[0].String(),
				Denom: "invalid denom",
			},
			0,
			false,
			"",
		},
		{
			"denom",
			&types.QueryDepositsRequest{
				Denom: "bnb",
			},
			2,
			false,
			"",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.giveName, func() {
			res, err := suite.queryServer.Deposits(sdk.WrapSDKContext(suite.ctx), tt.giveRequest)

			if tt.shouldError {
				suite.Error(err)
				suite.Contains(err.Error(), tt.errorSubstr)
			} else {
				suite.NoError(err)
				suite.Equal(tt.wantDepositCounts, len(res.Deposits))
			}
		})
	}
}

func (suite *grpcQueryTestSuite) TestGrpcQueryTotalSupply() {
	testCases := []struct {
		name           string
		deposits       types.Deposits
		expectedSupply sdk.Coins
	}{
		{
			name:           "returns zeros when there's no supply",
			deposits:       []types.Deposit{},
			expectedSupply: sdk.NewCoins(),
		},
		{
			name: "returns supply of one denom deposited from multiple accounts",
			deposits: []types.Deposit{
				dep(suite.addrs[0], sdk.NewCoins(c("busd", 1e6))),
				dep(suite.addrs[1], sdk.NewCoins(c("busd", 1e6))),
			},
			expectedSupply: sdk.NewCoins(c("busd", 2e6)),
		},
		{
			name: "returns supply of multiple denoms deposited from single account",
			deposits: []types.Deposit{
				dep(suite.addrs[0], sdk.NewCoins(c("busd", 1e6), c("bnb", 1e6))),
			},
			expectedSupply: sdk.NewCoins(c("busd", 1e6), c("bnb", 1e6)),
		},
		{
			name: "returns supply of multiple denoms deposited from multiple accounts",
			deposits: []types.Deposit{
				dep(suite.addrs[0], sdk.NewCoins(c("busd", 1e6), c("bnb", 1e6))),
				dep(suite.addrs[1], sdk.NewCoins(c("busd", 1e6), c("bnb", 1e6))),
			},
			expectedSupply: sdk.NewCoins(c("busd", 2e6), c("bnb", 2e6)),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			// setup deposits
			suite.addDeposits(tc.deposits)

			res, err := suite.queryServer.TotalSupply(
				sdk.WrapSDKContext(suite.ctx),
				&types.QueryTotalSupplyRequest{},
			)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedSupply, res.Result)
		})
	}

	suite.Run("aggregates bkava denoms, accounting for slashing", func() {
		suite.SetupTest()

		address1, derivatives1, _ := suite.createAccountWithDerivatives(bkava1, sdk.NewInt(1e9))
		address2, derivatives2, _ := suite.createAccountWithDerivatives(bkava2, sdk.NewInt(1e9))

		// bond validators
		staking.EndBlocker(suite.ctx, suite.tApp.GetStakingKeeper())
		// slash val2 - its shares are now 80% as valuable!
		err := suite.slashValidator(sdk.ValAddress(address2), sdk.MustNewDecFromStr("0.2"))
		suite.Require().NoError(err)

		suite.addDeposits(
			types.Deposits{
				dep(address1, cs(derivatives1)),
				dep(address2, cs(derivatives2)),
			},
		)

		expectedSupply := sdk.NewCoins(
			sdk.NewCoin(
				"bkava",
				sdk.NewIntFromUint64(1e9). // derivative 1
								Add(sdk.NewInt(1e9).MulRaw(80).QuoRaw(100))), // derivative 2: original value * 80%
		)

		res, err := suite.queryServer.TotalSupply(
			sdk.WrapSDKContext(suite.ctx),
			&types.QueryTotalSupplyRequest{},
		)
		suite.Require().NoError(err)
		suite.Require().Equal(expectedSupply, res.Result)
	})
}

func (suite *grpcQueryTestSuite) addDeposits(deposits types.Deposits) {
	for _, dep := range deposits {
		suite.NotPanics(func() {
			err := suite.keeper.Deposit(suite.ctx, dep.Depositor, dep.Amount)
			suite.Require().NoError(err)
		})
	}
}

// createUnbondedValidator creates an unbonded validator with the given amount of self-delegation.
func (suite *grpcQueryTestSuite) createUnbondedValidator(address sdk.ValAddress, selfDelegation sdk.Coin, minSelfDelegation sdk.Int) error {
	msg, err := stakingtypes.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		minSelfDelegation,
	)
	if err != nil {
		return err
	}

	msgServer := stakingkeeper.NewMsgServerImpl(suite.tApp.GetStakingKeeper())
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(suite.ctx), msg)
	return err
}

// createAccountWithDerivatives creates an account with the given amount and denom of derivative token.
// Internally, it creates a validator account and mints derivatives from the validator's self delegation.
func (suite *grpcQueryTestSuite) createAccountWithDerivatives(denom string, amount sdk.Int) (sdk.AccAddress, sdk.Coin, sdk.Coins) {
	bondDenom := suite.tApp.GetStakingKeeper().BondDenom(suite.ctx)
	valAddress, err := liquidtypes.ParseLiquidStakingTokenDenom(denom)
	suite.Require().NoError(err)
	address := sdk.AccAddress(valAddress)

	remainingSelfDelegation := sdk.NewInt(1e6)
	selfDelegation := sdk.NewCoin(
		bondDenom,
		amount.Add(remainingSelfDelegation),
	)

	// create & fund account
	// ak := suite.tApp.GetAccountKeeper()
	// acc := ak.NewAccountWithAddress(suite.ctx, address)
	// ak.SetAccount(suite.ctx, acc)
	err = suite.tApp.FundAccount(suite.ctx, address, sdk.NewCoins(selfDelegation))
	suite.Require().NoError(err)

	err = suite.createUnbondedValidator(valAddress, selfDelegation, remainingSelfDelegation)
	suite.Require().NoError(err)

	toConvert := sdk.NewCoin(bondDenom, amount)
	derivatives, err := suite.tApp.GetLiquidKeeper().MintDerivative(suite.ctx,
		address,
		valAddress,
		toConvert,
	)
	suite.Require().NoError(err)

	fullBalance := suite.tApp.GetBankKeeper().GetAllBalances(suite.ctx, address)

	return address, derivatives, fullBalance
}

// slashValidator slashes the validator with the given address by the given percentage.
func (suite *grpcQueryTestSuite) slashValidator(address sdk.ValAddress, slashFraction sdk.Dec) error {
	stakingKeeper := suite.tApp.GetStakingKeeper()

	validator, found := stakingKeeper.GetValidator(suite.ctx, address)
	suite.Require().True(found)
	consAddr, err := validator.GetConsAddr()
	suite.Require().NoError(err)

	// Assume infraction was at current height. Note unbonding delegations and redelegations are only slashed if created after
	// the infraction height so none will be slashed.
	infractionHeight := suite.ctx.BlockHeight()

	power := stakingKeeper.TokensToConsensusPower(suite.ctx, validator.GetTokens())

	stakingKeeper.Slash(suite.ctx, consAddr, infractionHeight, power, slashFraction)
	return nil
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
