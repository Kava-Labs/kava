package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquidstaking/keeper"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite
	App           app.TestApp
	Ctx           sdk.Context
	Keeper        keeper.Keeper
	BankKeeper    bankkeeper.Keeper
	StakingKeeper stakingkeeper.Keeper
}

// The default state used by each test
func (suite *KeeperTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	gen := app.NewAuthBankGenesisBuilder().BuildMarshalled(tApp.AppCodec())
	tApp.InitializeFromGenesisStates(gen)

	suite.App = tApp
	suite.Ctx = ctx
	suite.Keeper = tApp.GetLiquidStakingKeeper()
	suite.StakingKeeper = tApp.GetStakingKeeper()
	suite.BankKeeper = tApp.GetBankKeeper()
}

// CreateAccount creates a new account from the provided balance
func (suite *KeeperTestSuite) CreateAccount(initialBalance sdk.Coins) authtypes.AccountI {
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	ak := suite.App.GetAccountKeeper()

	acc := ak.NewAccountWithAddress(suite.Ctx, addrs[0])
	ak.SetAccount(suite.Ctx, acc)

	err := simapp.FundAccount(suite.BankKeeper, suite.Ctx, acc.GetAddress(), initialBalance)
	suite.Require().NoError(err)

	return acc
}

// CreateVestingAccount creates a new vesting account from the provided balance and vesting balance
func (suite *KeeperTestSuite) CreateVestingAccount(initialBalance sdk.Coins, vestingBalance sdk.Coins) authtypes.AccountI {
	acc := suite.CreateAccount(initialBalance)
	bacc := acc.(*authtypes.BaseAccount)

	periods := vestingtypes.Periods{
		vestingtypes.Period{
			Length: 31556952,
			Amount: vestingBalance,
		},
	}
	vacc := vestingtypes.NewPeriodicVestingAccount(bacc, initialBalance, time.Now().Unix(), periods) // TODO is initialBalance correct for originalVesting?

	return vacc
}

func (suite *KeeperTestSuite) deliverMsgCreateValidator(ctx sdk.Context, address sdk.ValAddress, selfDelegation sdk.Coin) error {
	msg, err := stakingtypes.NewMsgCreateValidator(
		address,
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		stakingtypes.Description{},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1_000_000),
	)
	if err != nil {
		return err
	}

	handleStakingMsg := staking.NewHandler(suite.StakingKeeper)
	_, err = handleStakingMsg(ctx, msg)
	return err
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
