package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	supplyExported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/kava-labs/kava/x/kavadist"
	validatorvesting "github.com/kava-labs/kava/x/validator-vesting"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestSendCoinsToPeriodicVestingAccount() {
	suite.setupChain()
	// insert a period into an existing vesting schedule
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], cs(c("ukava", 100)), 5)
	suite.NoError(err)
	acc := suite.getAccount(suite.addrs[0])
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	expectedPeriods := vesting.Periods{
		vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 500)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 500)), vacc.Coins)
	suite.Equal(int64(116), vacc.EndTime)

	// append a period to the end of an existing vesting schedule
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], cs(c("ukava", 100)), 17)
	suite.NoError(err)
	acc = suite.getAccount(suite.addrs[0])
	vacc, ok = acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	expectedPeriods = append(expectedPeriods, vesting.Period{1, cs(c("ukava", 100))})
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 600)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 600)), vacc.Coins)
	suite.Equal(int64(117), vacc.EndTime)

	// append a period to the end of a completed vesting schedule
	suite.ctx = suite.ctx.WithBlockTime(time.Unix(120, 0))
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], cs(c("ukava", 100)), 5)
	suite.NoError(err)
	acc = suite.getAccount(suite.addrs[0])
	vacc, ok = acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	expectedPeriods = append(expectedPeriods, vesting.Period{8, cs(c("ukava", 100))})
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 700)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 700)), vacc.Coins)
	suite.Equal(int64(125), vacc.EndTime)

	// prepend a period to an upcoming vesting schedule
	suite.ctx = suite.ctx.WithBlockTime(time.Unix(90, 0))
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], cs(c("ukava", 100)), 5)
	suite.NoError(err)
	expectedPeriods = vesting.Periods{
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
	}
	acc = suite.getAccount(suite.addrs[0])
	vacc, ok = acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 800)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 800)), vacc.Coins)
	suite.Equal(int64(125), vacc.EndTime)
	suite.Equal(int64(90), vacc.StartTime)

	// add a period that coincides with an existing end time
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], cs(c("ukava", 100)), 11)
	suite.NoError(err)
	expectedPeriods = vesting.Periods{
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(6), Amount: cs(c("ukava", 200))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(6), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
	}
	acc = suite.getAccount(suite.addrs[0])
	vacc, ok = acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 900)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 900)), vacc.Coins)
	suite.Equal(int64(125), vacc.EndTime)
	suite.Equal(int64(90), vacc.StartTime)

	// sending coins from empty module account errors
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[0], cs(c("ukava", 100)), 11)
	suite.Equal(types.CodeInsufficientBalance, err.Result().Code)
}

func (suite *KeeperTestSuite) TestSendCoinsToBaseAccount() {
	suite.setupChain()
	// send coins to base account
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[1], cs(c("ukava", 100)), 5)
	suite.NoError(err)
	acc := suite.getAccount(suite.addrs[1])
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	expectedPeriods := vesting.Periods{
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	suite.Equal(expectedPeriods, vacc.VestingPeriods)
	suite.Equal(cs(c("ukava", 100)), vacc.OriginalVesting)
	suite.Equal(cs(c("ukava", 500)), vacc.Coins)
	suite.Equal(int64(105), vacc.EndTime)
	suite.Equal(int64(100), vacc.StartTime)

}

func (suite *KeeperTestSuite) TestSendCoinsToInvalidAccount() {
	suite.setupChain()
	err := suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, suite.addrs[2], cs(c("ukava", 100)), 5)
	suite.Equal(types.CodeInvalidAccountType, err.Result().Code)
	macc := suite.getModuleAccount(cdp.ModuleName)
	err = suite.keeper.SendTimeLockedCoinsToAccount(suite.ctx, kavadist.ModuleName, macc.GetAddress(), cs(c("ukava", 100)), 5)
	suite.Equal(types.CodeInvalidAccountType, err.Result().Code)
}

func (suite *KeeperTestSuite) TestPayoutClaim() {
	suite.setupClaims()
	err := suite.keeper.PayoutClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.NoError(err)
	acc := suite.getAccount(suite.addrs[0])
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	suite.Equal(cs(c("ukava", 500)), vacc.OriginalVesting)

	err = suite.keeper.PayoutClaim(suite.ctx, suite.addrs[1], "bnb", 1)
	suite.NoError(err)
	acc = suite.getAccount(suite.addrs[1])
	vacc, ok = acc.(*vesting.PeriodicVestingAccount)
	suite.True(ok)
	suite.Equal(cs(c("ukava", 100)), vacc.OriginalVesting)

	err = suite.keeper.PayoutClaim(suite.ctx, suite.addrs[3], "bnb", 1)
	suite.Equal(types.CodeClaimNotFound, err.Result().Code)
	err = suite.keeper.PayoutClaim(suite.ctx, suite.addrs[0], "xrp", 1)
	suite.Equal(types.CodeClaimPeriodNotFound, err.Result().Code)
}

func (suite *KeeperTestSuite) TestDeleteExpiredClaimPeriods() {
	suite.setupExpiredClaims()
	_, found := suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.True(found)
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "xrp")
	suite.True(found)
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.True(found)
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "xrp", 1)
	suite.True(found)
	suite.NotPanics(func() {
		suite.keeper.DeleteExpiredClaimsAndClaimPeriods(suite.ctx)
	})
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "bnb")
	suite.False(found)
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "bnb", 1)
	suite.False(found)
	_, found = suite.keeper.GetClaimPeriod(suite.ctx, 1, "xrp")
	suite.True(found)
	_, found = suite.keeper.GetClaim(suite.ctx, suite.addrs[0], "xrp", 1)
	suite.True(found)

}

func (suite *KeeperTestSuite) setupChain() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
		})
	tApp.InitializeFromGenesisStates(
		authGS,
	)
	supplyKeeper := tApp.GetSupplyKeeper()
	macc := supplyKeeper.GetModuleAccount(ctx, kavadist.ModuleName)
	err := supplyKeeper.MintCoins(ctx, macc.GetName(), cs(c("ukava", 500)))
	suite.NoError(err)

	// add a periodic vesting account
	ak := tApp.GetAccountKeeper()
	acc := ak.GetAccount(ctx, addrs[0])
	bacc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	periods := vesting.Periods{
		vesting.Period{Length: int64(1), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(2), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(8), Amount: cs(c("ukava", 100))},
		vesting.Period{Length: int64(5), Amount: cs(c("ukava", 100))},
	}
	bva, err2 := vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
	suite.NoError(err2)
	pva := vesting.NewPeriodicVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods)
	ak.SetAccount(ctx, pva)

	// add a validator vesting account
	acc = ak.GetAccount(ctx, addrs[2])
	bacc = auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	bva, err2 = vesting.NewBaseVestingAccount(bacc, cs(c("ukava", 400)), ctx.BlockTime().Unix()+16)
	suite.NoError(err2)
	vva := validatorvesting.NewValidatorVestingAccountRaw(bva, ctx.BlockTime().Unix(), periods, sdk.ConsAddress{}, nil, 90)
	ak.SetAccount(ctx, vva)
	suite.app = tApp
	suite.keeper = tApp.GetIncentiveKeeper()
	suite.ctx = ctx
	suite.addrs = addrs
}

func (suite *KeeperTestSuite) setupClaims() {
	suite.setupChain()
	cp1 := types.NewClaimPeriod("bnb", 1, suite.ctx.BlockTime().Add(time.Hour*168), time.Hour*8766)
	suite.keeper.SetClaimPeriod(suite.ctx, cp1)
	c1 := types.NewClaim(suite.addrs[0], c("ukava", 100), "bnb", 1)
	c2 := types.NewClaim(suite.addrs[0], c("ukava", 100), "xrp", 1)
	c3 := types.NewClaim(suite.addrs[1], c("ukava", 100), "bnb", 1)
	suite.keeper.SetClaim(suite.ctx, c1)
	suite.keeper.SetClaim(suite.ctx, c2)
	suite.keeper.SetClaim(suite.ctx, c3)
}

func (suite *KeeperTestSuite) getAccount(addr sdk.AccAddress) exported.Account {
	ak := suite.app.GetAccountKeeper()
	return ak.GetAccount(suite.ctx, addr)
}

func (suite *KeeperTestSuite) getModuleAccount(name string) supplyExported.ModuleAccountI {
	sk := suite.app.GetSupplyKeeper()
	return sk.GetModuleAccount(suite.ctx, name)
}

func (suite *KeeperTestSuite) setupExpiredClaims() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: time.Unix(100, 0)})
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	authGS := app.NewAuthGenState(
		addrs,
		[]sdk.Coins{
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
			cs(c("ukava", 400)),
		})
	tApp.InitializeFromGenesisStates(
		authGS,
	)
	cp1 := types.NewClaimPeriod("bnb", 1, time.Unix(90, 0), time.Hour*8766)
	cp2 := types.NewClaimPeriod("xrp", 1, time.Unix(110, 0), time.Hour*8766)
	suite.keeper = tApp.GetIncentiveKeeper()
	suite.keeper.SetClaimPeriod(ctx, cp1)
	suite.keeper.SetClaimPeriod(ctx, cp2)
	c1 := types.NewClaim(addrs[0], c("ukava", 1000000), "bnb", 1)
	c2 := types.NewClaim(addrs[0], c("ukava", 1000000), "xrp", 1)
	suite.keeper.SetClaim(ctx, c1)
	suite.keeper.SetClaim(ctx, c2)
	suite.app = tApp
	suite.ctx = ctx
	suite.addrs = addrs
}

func stringPva(pva *vesting.PeriodicVestingAccount) string {
	return fmt.Sprintf(`Periodic Vesting Account:
	Address: %s
	Coins: %s
	Original Vesting: %s
	Start: %d,
	End: %d,
	Vesting Periods: %s
	AccountNumber: %d
	Sequence: %d`,
		pva.GetAddress(), pva.GetCoins(), pva.GetOriginalVesting(), pva.StartTime, pva.EndTime, (pva.VestingPeriods), pva.GetAccountNumber(), pva.GetSequence())
}

func stringPeriod(p vesting.Period) string {
	return fmt.Sprintf("\t\tLength: %d, Amount: %s", p.Length, p.Amount)
}

func stringPeriods(periods vesting.Periods) string {
	out := ""
	for _, p := range periods {
		out += fmt.Sprintf("\n%s", stringPeriod(p))
	}
	return out
}
