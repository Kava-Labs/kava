package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kava-labs/kava/x/earn/keeper"
	"github.com/kava-labs/kava/x/earn/testutil"
	"github.com/kava-labs/kava/x/earn/types"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
)

type grpcQueryTestSuite struct {
	testutil.Suite

	queryClient types.QueryClient
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.Keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}

func (suite *grpcQueryTestSuite) TestQueryParams() {
	vaultDenom := "usdx"

	res, err := suite.queryClient.Params(context.Background(), types.NewQueryParamsRequest())
	suite.Require().NoError(err)
	// ElementsMatch instead of Equal because AllowedVaults{} != AllowedVaults(nil)
	suite.Require().ElementsMatch(types.DefaultParams().AllowedVaults, res.Params.AllowedVaults)

	// Add vault to params
	suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)

	// Query again for added vault
	res, err = suite.queryClient.Params(context.Background(), types.NewQueryParamsRequest())
	suite.Require().NoError(err)
	suite.Require().Equal(
		types.AllowedVaults{
			types.NewAllowedVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil),
		},
		res.Params.AllowedVaults,
	)
}

func (suite *grpcQueryTestSuite) TestVaults_ZeroSupply() {
	// Add vaults
	suite.CreateVault("usdx", types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("busd", types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)

	suite.Run("single", func() {
		res, err := suite.queryClient.Vault(context.Background(), types.NewQueryVaultRequest("usdx"))
		suite.Require().NoError(err)
		suite.Require().Equal(
			types.VaultResponse{
				Denom:             "usdx",
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       sdk.NewDec(0).String(),
				TotalValue:        sdk.NewInt(0),
			},
			res.Vault,
		)
	})

	suite.Run("all", func() {
		res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest())
		suite.Require().NoError(err)
		suite.Require().ElementsMatch([]types.VaultResponse{
			{
				Denom:             "usdx",
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       sdk.ZeroDec().String(),
				TotalValue:        sdk.ZeroInt(),
			},
			{
				Denom:             "busd",
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       sdk.ZeroDec().String(),
				TotalValue:        sdk.ZeroInt(),
			},
		},
			res.Vaults,
		)
	})
}

func (suite *grpcQueryTestSuite) TestVaults_WithSupply() {
	vaultDenom := "usdx"
	vault2Denom := testutil.TestBkavaDenoms[0]

	depositAmount := sdk.NewInt64Coin(vaultDenom, 100)
	deposit2Amount := sdk.NewInt64Coin(vault2Denom, 100)

	suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("bkava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(
		sdk.NewInt64Coin(vaultDenom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
	), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), deposit2Amount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest())
	suite.Require().NoError(err)
	suite.Require().Len(res.Vaults, 2)
	suite.Require().ElementsMatch(
		[]types.VaultResponse{
			{
				Denom:             vaultDenom,
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       depositAmount.Amount.ToDec().String(),
				TotalValue:        depositAmount.Amount,
			},
			{
				Denom:             vault2Denom,
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_SAVINGS},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       deposit2Amount.Amount.ToDec().String(),
				TotalValue:        deposit2Amount.Amount,
			},
		},
		res.Vaults,
	)
}

func (suite *grpcQueryTestSuite) TestVaults_MixedSupply() {
	vaultDenom := "usdx"
	vault2Denom := "busd"
	vault3Denom := testutil.TestBkavaDenoms[0]

	depositAmount := sdk.NewInt64Coin(vault3Denom, 100)

	suite.CreateVault(vaultDenom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault(vault2Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("bkava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	acc := suite.CreateAccount(sdk.NewCoins(
		sdk.NewInt64Coin(vaultDenom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
		sdk.NewInt64Coin(vault3Denom, 1000),
	), 0)

	err := suite.Keeper.Deposit(suite.Ctx, acc.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	res, err := suite.queryClient.Vaults(context.Background(), types.NewQueryVaultsRequest())
	suite.Require().NoError(err)
	suite.Require().Len(res.Vaults, 3)
	suite.Require().ElementsMatch(
		[]types.VaultResponse{
			{
				Denom:             vaultDenom,
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       sdk.ZeroDec().String(),
				TotalValue:        sdk.ZeroInt(),
			},
			{
				Denom:             vault2Denom,
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_HARD},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       sdk.ZeroDec().String(),
				TotalValue:        sdk.ZeroInt(),
			},
			{
				Denom:             vault3Denom,
				Strategies:        []types.StrategyType{types.STRATEGY_TYPE_SAVINGS},
				IsPrivateVault:    false,
				AllowedDepositors: nil,
				TotalShares:       depositAmount.Amount.ToDec().String(),
				TotalValue:        depositAmount.Amount,
			},
		},
		res.Vaults,
	)
}

func (suite *grpcQueryTestSuite) TestVault_NotFound() {
	_, err := suite.queryClient.Vault(context.Background(), types.NewQueryVaultRequest("usdx"))
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Errorf(codes.NotFound, "vault not found with specified denom"))
}

func (suite *grpcQueryTestSuite) TestDeposits() {
	vault1Denom := "usdx"
	vault2Denom := "busd"
	vault3Denom := testutil.TestBkavaDenoms[0]

	// Add vaults
	suite.CreateVault(vault1Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault(vault2Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("bkava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	startBalance := sdk.NewCoins(
		sdk.NewInt64Coin(vault1Denom, 1000),
		sdk.NewInt64Coin(vault2Denom, 1000),
		sdk.NewInt64Coin(vault3Denom, 1000),
	)

	deposit1Amount := sdk.NewInt64Coin(vault1Denom, 100)
	deposit2Amount := sdk.NewInt64Coin(vault2Denom, 200)
	deposit3Amount := sdk.NewInt64Coin(vault3Denom, 200)

	// Accounts
	acc1 := suite.CreateAccount(startBalance, 0).GetAddress()
	acc2 := suite.CreateAccount(startBalance, 1).GetAddress()

	// Deposit into each vault from each account - 4 total deposits
	// Acc 1: usdx + busd
	// Acc 2: usdx + bkava
	err := suite.Keeper.Deposit(suite.Ctx, acc1, deposit1Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)
	err = suite.Keeper.Deposit(suite.Ctx, acc1, deposit2Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, acc2, deposit1Amount, types.STRATEGY_TYPE_HARD)
	suite.Require().NoError(err)
	err = suite.Keeper.Deposit(suite.Ctx, acc2, deposit3Amount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	suite.Run("specific vault", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), vault1Denom, nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatchf(
			[]types.DepositResponse{
				{
					Depositor: acc1.String(),
					// Only includes specified deposit shares
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit1Amount.Denom, deposit1Amount.Amount.ToDec()),
					),
					// Only the specified vault denom value
					Value: sdk.NewCoins(deposit1Amount),
				},
			},
			res.Deposits,
			"deposits should match, got %v",
			res.Deposits,
		)
	})

	suite.Run("specific bkava vault", func() {
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc2.String(), vault3Denom, nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatchf(
			[]types.DepositResponse{
				{
					Depositor: acc2.String(),
					// Only includes specified deposit shares
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit3Amount.Denom, deposit3Amount.Amount.ToDec()),
					),
					// Only the specified vault denom value
					Value: sdk.NewCoins(deposit3Amount),
				},
			},
			res.Deposits,
			"deposits should match, got %v",
			res.Deposits,
		)
	})

	suite.Run("invalid vault", func() {
		_, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), "notavaliddenom", nil),
		)
		suite.Require().Error(err)
		suite.Require().ErrorIs(err, status.Errorf(codes.NotFound, "vault for notavaliddenom not found"))
	})

	suite.Run("all vaults", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), "", nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatch(
			[]types.DepositResponse{
				{
					Depositor: acc1.String(),
					Shares: types.NewVaultShares(
						types.NewVaultShare(deposit1Amount.Denom, deposit1Amount.Amount.ToDec()),
						types.NewVaultShare(deposit2Amount.Denom, deposit2Amount.Amount.ToDec()),
					),
					Value: sdk.NewCoins(deposit1Amount, deposit2Amount),
				},
			},
			res.Deposits,
		)
	})
}

func (suite *grpcQueryTestSuite) TestDeposits_NoDeposits() {
	vault1Denom := "usdx"
	vault2Denom := "busd"

	// Add vaults
	suite.CreateVault(vault1Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault(vault2Denom, types.StrategyTypes{types.STRATEGY_TYPE_HARD}, false, nil)
	suite.CreateVault("bkava", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)

	// Accounts
	acc1 := suite.CreateAccount(sdk.NewCoins(), 0).GetAddress()

	suite.Run("specific vault", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), vault1Denom, nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatchf(
			[]types.DepositResponse{
				{
					Depositor: acc1.String(),
					// Zero shares and zero value
					Shares: nil,
					Value:  nil,
				},
			},
			res.Deposits,
			"deposits should match, got %v",
			res.Deposits,
		)
	})

	suite.Run("all vaults", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(acc1.String(), "", nil),
		)
		suite.Require().NoError(err)
		suite.Require().Empty(res.Deposits)
	})
}

func (suite *grpcQueryTestSuite) TestDeposits_NoDepositor() {
	_, err := suite.queryClient.Deposits(
		context.Background(),
		types.NewQueryDepositsRequest("", "usdx", nil),
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "depositor is required"))
}

func (suite *grpcQueryTestSuite) TestDeposits_InvalidAddress() {
	_, err := suite.queryClient.Deposits(
		context.Background(),
		types.NewQueryDepositsRequest("asdf", "usdx", nil),
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "Invalid address"))

	_, err = suite.queryClient.Deposits(
		context.Background(),
		types.NewQueryDepositsRequest("asdf", "", nil),
	)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, status.Error(codes.InvalidArgument, "Invalid address"))
}

func (suite *grpcQueryTestSuite) TestDeposits_bKava() {
	// vault denom is only "bkava" which has it's own special handler
	suite.CreateVault(
		"bkava",
		types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS},
		false,
		[]sdk.AccAddress{},
	)

	suite.CreateVault(
		"ukava",
		types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS},
		false,
		[]sdk.AccAddress{},
	)

	address1, derivatives1, _ := suite.createAccountWithDerivatives(testutil.TestBkavaDenoms[0], sdk.NewInt(1e9))
	address2, derivatives2, _ := suite.createAccountWithDerivatives(testutil.TestBkavaDenoms[1], sdk.NewInt(1e9))

	err := suite.App.FundAccount(suite.Ctx, address1, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1e9))))
	suite.Require().NoError(err)

	// Slash the last validator to reduce the value of it's derivatives to test bkava to underlying token conversion.
	// First call end block to bond validator to enable slashing.
	staking.EndBlocker(suite.Ctx, suite.App.GetStakingKeeper())
	err = suite.slashValidator(sdk.ValAddress(address2), sdk.MustNewDecFromStr("0.5"))
	suite.Require().NoError(err)

	suite.Run("no deposits", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(address1.String(), "bkava", nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		suite.Require().ElementsMatchf(
			[]types.DepositResponse{
				{
					Depositor: address1.String(),
					// Zero shares for "bkava" aggregate
					Shares: nil,
					// Only the specified vault denom value
					Value: nil,
				},
			},
			res.Deposits,
			"deposits should match, got %v",
			res.Deposits,
		)
	})

	err = suite.Keeper.Deposit(suite.Ctx, address1, derivatives1, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	err = suite.BankKeeper.SendCoins(suite.Ctx, address2, address1, sdk.NewCoins(derivatives2))
	suite.Require().NoError(err)
	err = suite.Keeper.Deposit(suite.Ctx, address1, derivatives2, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, address1, sdk.NewInt64Coin("ukava", 1e6), types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	suite.Run("multiple deposits", func() {
		// Query all deposits for account 1
		res, err := suite.queryClient.Deposits(
			context.Background(),
			types.NewQueryDepositsRequest(address1.String(), "bkava", nil),
		)
		suite.Require().NoError(err)
		suite.Require().Len(res.Deposits, 1)
		// first validator isn't slashed, so bkava units equal to underlying staked tokens
		// last validator slashed 50% so derivatives are worth half
		// Excludes non-bkava deposits
		expectedValue := derivatives1.Amount.Add(derivatives2.Amount.QuoRaw(2))
		suite.Require().ElementsMatchf(
			[]types.DepositResponse{
				{
					Depositor: address1.String(),
					// Zero shares for "bkava" aggregate
					Shares: nil,
					// Value returned in units of staked token
					Value: sdk.NewCoins(
						sdk.NewCoin(suite.bondDenom(), expectedValue),
					),
				},
			},
			res.Deposits,
			"deposits should match, got %v",
			res.Deposits,
		)
	})
}

func (suite *grpcQueryTestSuite) TestVault_bKava_Single() {
	vaultDenom := "bkava"
	coinDenom := testutil.TestBkavaDenoms[0]

	startBalance := sdk.NewInt64Coin(coinDenom, 1000)
	depositAmount := sdk.NewInt64Coin(coinDenom, 100)

	acc1 := suite.CreateAccount(sdk.NewCoins(startBalance), 0)

	// vault denom is only "bkava" which has it's own special handler
	suite.CreateVault(
		vaultDenom,
		types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS},
		false,
		[]sdk.AccAddress{},
	)

	err := suite.Keeper.Deposit(suite.Ctx, acc1.GetAddress(), depositAmount, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(
		err,
		"should be able to deposit bkava derivative denom in bkava vault",
	)

	res, err := suite.queryClient.Vault(
		context.Background(),
		types.NewQueryVaultRequest(coinDenom),
	)
	suite.Require().NoError(err)
	suite.Require().Equal(
		types.VaultResponse{
			Denom: coinDenom,
			Strategies: types.StrategyTypes{
				types.STRATEGY_TYPE_SAVINGS,
			},
			IsPrivateVault:    false,
			AllowedDepositors: []string(nil),
			TotalShares:       "100.000000000000000000",
			TotalValue:        sdk.NewInt(100),
		},
		res.Vault,
	)
}

func (suite *grpcQueryTestSuite) TestVault_bKava_Aggregate() {
	vaultDenom := "bkava"

	address1, derivatives1, _ := suite.createAccountWithDerivatives(testutil.TestBkavaDenoms[0], sdk.NewInt(1e9))
	address2, derivatives2, _ := suite.createAccountWithDerivatives(testutil.TestBkavaDenoms[1], sdk.NewInt(1e9))
	address3, derivatives3, _ := suite.createAccountWithDerivatives(testutil.TestBkavaDenoms[2], sdk.NewInt(1e9))
	// Slash the last validator to reduce the value of it's derivatives to test bkava to underlying token conversion.
	// First call end block to bond validator to enable slashing.
	staking.EndBlocker(suite.Ctx, suite.App.GetStakingKeeper())
	err := suite.slashValidator(sdk.ValAddress(address3), sdk.MustNewDecFromStr("0.5"))
	suite.Require().NoError(err)

	// vault denom is only "bkava" which has it's own special handler
	suite.CreateVault(
		vaultDenom,
		types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS},
		false,
		[]sdk.AccAddress{},
	)

	err = suite.Keeper.Deposit(suite.Ctx, address1, derivatives1, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, address2, derivatives2, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	err = suite.Keeper.Deposit(suite.Ctx, address3, derivatives3, types.STRATEGY_TYPE_SAVINGS)
	suite.Require().NoError(err)

	// Query "bkava" to get aggregate amount
	res, err := suite.queryClient.Vault(
		context.Background(),
		types.NewQueryVaultRequest(vaultDenom),
	)
	suite.Require().NoError(err)
	// first two validators are not slashed, so bkava units equal to underlying staked tokens
	expectedValue := derivatives1.Amount.Add(derivatives2.Amount)
	// last validator slashed 50% so derivatives are worth half
	expectedValue = expectedValue.Add(derivatives2.Amount.QuoRaw(2))
	suite.Require().Equal(
		types.VaultResponse{
			Denom: vaultDenom,
			Strategies: types.StrategyTypes{
				types.STRATEGY_TYPE_SAVINGS,
			},
			IsPrivateVault:    false,
			AllowedDepositors: []string(nil),
			// No shares for aggregate
			TotalShares: "0",
			TotalValue:  expectedValue,
		},
		res.Vault,
	)
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

	msgServer := stakingkeeper.NewMsgServerImpl(suite.App.GetStakingKeeper())
	_, err = msgServer.CreateValidator(sdk.WrapSDKContext(suite.Ctx), msg)
	return err
}

// createAccountWithDerivatives creates an account with the given amount and denom of derivative token.
// Internally, it creates a validator account and mints derivatives from the validator's self delegation.
func (suite *grpcQueryTestSuite) createAccountWithDerivatives(denom string, amount sdk.Int) (sdk.AccAddress, sdk.Coin, sdk.Coins) {
	valAddress, err := liquidtypes.ParseLiquidStakingTokenDenom(denom)
	suite.Require().NoError(err)
	address := sdk.AccAddress(valAddress)

	remainingSelfDelegation := sdk.NewInt(1e6)
	selfDelegation := sdk.NewCoin(
		suite.bondDenom(),
		amount.Add(remainingSelfDelegation),
	)

	suite.NewAccountFromAddr(address, sdk.NewCoins(selfDelegation))

	err = suite.createUnbondedValidator(valAddress, selfDelegation, remainingSelfDelegation)
	suite.Require().NoError(err)

	toConvert := sdk.NewCoin(suite.bondDenom(), amount)
	derivatives, err := suite.App.GetLiquidKeeper().MintDerivative(suite.Ctx,
		address,
		valAddress,
		toConvert,
	)
	suite.Require().NoError(err)

	fullBalance := suite.BankKeeper.GetAllBalances(suite.Ctx, address)

	return address, derivatives, fullBalance
}

// slashValidator slashes the validator with the given address by the given percentage.
func (suite *grpcQueryTestSuite) slashValidator(address sdk.ValAddress, slashFraction sdk.Dec) error {
	stakingKeeper := suite.App.GetStakingKeeper()

	validator, found := stakingKeeper.GetValidator(suite.Ctx, address)
	suite.Require().True(found)
	consAddr, err := validator.GetConsAddr()
	suite.Require().NoError(err)

	// Assume infraction was at current height. Note unbonding delegations and redelegations are only slashed if created after
	// the infraction height so none will be slashed.
	infractionHeight := suite.Ctx.BlockHeight()

	power := stakingKeeper.TokensToConsensusPower(suite.Ctx, validator.GetTokens())

	stakingKeeper.Slash(suite.Ctx, consAddr, infractionHeight, power, slashFraction)
	return nil
}

// bondDenom fetches the staking denom from the staking module.
func (suite *grpcQueryTestSuite) bondDenom() string {
	return suite.App.GetStakingKeeper().BondDenom(suite.Ctx)
}
