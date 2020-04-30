package validatorvesting

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"

	"github.com/kava-labs/kava/x/validator-vesting/keeper"
	"github.com/kava-labs/kava/x/validator-vesting/types"
)

func TestBeginBlockerZeroHeight(t *testing.T) {
	ctx, ak, _, stakingKeeper, _, vvk := keeper.CreateTestInput(t, false, 1000)
	now := tmtime.Now()
	vva := keeper.ValidatorVestingDelegatorTestAccount(now)
	ak.SetAccount(ctx, vva)
	delTokens := sdk.TokensFromConsensusPower(30)
	vvk.SetValidatorVestingAccountKey(ctx, vva.Address)

	keeper.CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})

	val1, found := stakingKeeper.GetValidator(ctx, keeper.ValOpAddr1)
	require.True(t, found)
	_, err := stakingKeeper.Delegate(ctx, vva.Address, delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	// require that there exists one delegation
	var delegations int
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})

	require.Equal(t, 1, delegations)

	val := abci.Validator{
		Address: val1.ConsPubKey.Address(),
		Power:   val1.ConsensusPower(),
	}

	vva.ValidatorAddress = val1.ConsAddress()
	ak.SetAccount(ctx, vva)

	height := int64(1)
	blockTime := now
	addHour := func(t time.Time) time.Time { return t.Add(1 * time.Hour) }

	header := abci.Header{Height: height, Time: addHour(blockTime)}
	ctx = ctx.WithBlockHeader(header)

	// mark the validator as absent
	req := abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       abci.Validator{},
				SignedLastBlock: false,
			}},
		},
	}

	BeginBlocker(ctx, req, vvk)

	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	// require missed block counter doesn't increment because there's no voting history
	require.Equal(t, types.CurrentPeriodProgress{0, 1}, vva.CurrentPeriodProgress)

	// mark the validator as having missed
	req = abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: false,
			}},
		},
	}

	BeginBlocker(ctx, req, vvk)

	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, types.CurrentPeriodProgress{0, 2}, vva.CurrentPeriodProgress)
}

func TestBeginBlockerSignedBlock(t *testing.T) {
	ctx, ak, _, stakingKeeper, _, vvk := keeper.CreateTestInput(t, false, 1000)
	now := tmtime.Now()

	vva := keeper.ValidatorVestingDelegatorTestAccount(now)

	ak.SetAccount(ctx, vva)
	delTokens := sdk.TokensFromConsensusPower(30)
	vvk.SetValidatorVestingAccountKey(ctx, vva.Address)

	keeper.CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})

	val1, found := stakingKeeper.GetValidator(ctx, keeper.ValOpAddr1)
	require.True(t, found)
	_, err := stakingKeeper.Delegate(ctx, vva.Address, delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	// require that there exists one delegation
	var delegations int
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})

	require.Equal(t, 1, delegations)

	val := abci.Validator{
		Address: val1.ConsPubKey.Address(),
		Power:   val1.ConsensusPower(),
	}

	vva.ValidatorAddress = val1.ConsAddress()
	ak.SetAccount(ctx, vva)

	height := int64(1)
	blockTime := now

	addHour := func(t time.Time) time.Time { return t.Add(1 * time.Hour) }

	header := abci.Header{Height: height, Time: addHour(blockTime)}
	ctx = ctx.WithBlockHeader(header)
	// mark the validator as having signed
	req := abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: true,
			}},
		},
	}

	BeginBlocker(ctx, req, vvk)
	height++
	blockTime = addHour(blockTime)
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, types.CurrentPeriodProgress{0, 1}, vva.CurrentPeriodProgress)

	header = abci.Header{Height: height, Time: addHour(blockTime)}
	// mark the validator as having signed
	ctx = ctx.WithBlockHeader(header)
	req = abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: true,
			}},
		},
	}

	BeginBlocker(ctx, req, vvk)
	height++
	blockTime = addHour(blockTime)
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, types.CurrentPeriodProgress{0, 2}, vva.CurrentPeriodProgress)

	header = abci.Header{Height: height, Time: addHour(blockTime)}
	ctx = ctx.WithBlockHeader(header)
	// mark the validator as having missed
	req = abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: false,
			}},
		},
	}

	BeginBlocker(ctx, req, vvk)
	height++
	blockTime = addHour(blockTime)
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, types.CurrentPeriodProgress{1, 3}, vva.CurrentPeriodProgress)

	header = abci.Header{Height: height, Time: addHour(blockTime)}
	ctx = ctx.WithBlockHeader(header)
	// mark the validator as being absent
	req = abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       abci.Validator{},
				SignedLastBlock: true,
			}},
		},
	}

	BeginBlocker(ctx, req, vvk)
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, types.CurrentPeriodProgress{2, 4}, vva.CurrentPeriodProgress)
}

func TestBeginBlockerSuccessfulPeriod(t *testing.T) {
	height := int64(1)
	now := tmtime.Now()
	blockTime := now
	numBlocks := int64(14)
	addHour := func(t time.Time) time.Time { return t.Add(1 * time.Hour) }
	ctx, ak, _, stakingKeeper, _, vvk := keeper.CreateTestInput(t, false, 1000)

	vva := keeper.ValidatorVestingDelegatorTestAccount(now)

	ak.SetAccount(ctx, vva)
	vvk.SetValidatorVestingAccountKey(ctx, vva.Address)

	keeper.CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})

	val1, found := stakingKeeper.GetValidator(ctx, keeper.ValOpAddr1)
	require.True(t, found)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	val := abci.Validator{
		Address: val1.ConsPubKey.Address(),
		Power:   val1.ConsensusPower(),
	}

	vva.ValidatorAddress = val1.ConsAddress()
	ak.SetAccount(ctx, vva)

	for ; height < numBlocks; height++ {
		header := abci.Header{Height: height, Time: addHour(blockTime)}
		// mark the validator as having signed
		req := abci.RequestBeginBlock{
			Header: header,
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{{
					Validator:       val,
					SignedLastBlock: true,
				}},
			},
		}
		ctx = ctx.WithBlockHeader(header)
		BeginBlocker(ctx, req, vvk)
		blockTime = addHour(blockTime)

		if height == 12 {
			vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
			// require that missing sign count is set back to zero after the period increments.
			require.Equal(t, types.CurrentPeriodProgress{0, 0}, vva.CurrentPeriodProgress)
		}

	}

	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	// t.Log(vva.MarshalYAML())
	require.Equal(t, []types.VestingProgress{{true, true}, {false, false}, {false, false}}, vva.VestingPeriodProgress)
}

func TestBeginBlockerUnsuccessfulPeriod(t *testing.T) {
	height := int64(1)
	now := tmtime.Now()
	blockTime := now
	numBlocks := int64(13)
	addHour := func(t time.Time) time.Time { return t.Add(1 * time.Hour) }

	ctx, ak, _, stakingKeeper, supplyKeeper, vvk := keeper.CreateTestInput(t, false, 1000)

	initialSupply := supplyKeeper.GetSupply(ctx).GetTotal()
	keeper.CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})

	vva := keeper.ValidatorVestingDelegatorTestAccount(now)

	ak.SetAccount(ctx, vva)
	// delegate all coins
	delTokens := sdk.TokensFromConsensusPower(60)
	vvk.SetValidatorVestingAccountKey(ctx, vva.Address)

	val1, found := stakingKeeper.GetValidator(ctx, keeper.ValOpAddr1)
	require.True(t, found)
	_, err := stakingKeeper.Delegate(ctx, vva.Address, delTokens, sdk.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, stakingKeeper)

	// note that delegation modifies the account's state!
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)

	val := abci.Validator{
		Address: val1.ConsPubKey.Address(),
		Power:   val1.ConsensusPower(),
	}

	vva.ValidatorAddress = val1.ConsAddress()
	ak.SetAccount(ctx, vva)

	// run one period's worth of blocks
	for ; height < numBlocks; height++ {
		header := abci.Header{Height: height, Time: addHour(blockTime)}
		// mark the validator as having missed
		req := abci.RequestBeginBlock{
			Header: header,
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{{
					Validator:       val,
					SignedLastBlock: false,
				}},
			},
		}
		ctx = ctx.WithBlockHeader(header)
		BeginBlocker(ctx, req, vvk)
		blockTime = addHour(blockTime)
		vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	}

	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	// check that the period was unsuccessful
	require.Equal(t, []types.VestingProgress{{true, false}, {false, false}, {false, false}}, vva.VestingPeriodProgress)
	// check that there is debt after the period.
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin("stake", 30000000)}, vva.DebtAfterFailedVesting)

	var delegations int
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})
	// require that all delegations were unbonded
	require.Equal(t, 0, delegations)

	// complete the unbonding period
	header := abci.Header{Height: height, Time: blockTime.Add(time.Hour * 2)}
	req := abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: false,
			}},
		},
	}
	ctx = ctx.WithBlockHeader(header)
	BeginBlocker(ctx, req, vvk)
	_ = staking.EndBlocker(ctx, stakingKeeper)

	header = abci.Header{Height: height, Time: blockTime.Add(time.Hour * 2)}
	req = abci.RequestBeginBlock{
		Header: header,
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: false,
			}},
		},
	}
	ctx = ctx.WithBlockHeader(header)
	BeginBlocker(ctx, req, vvk)
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	// require that debt has reset to zero and coins balance is reduced by period 1 amount.
	require.Equal(t, vva.GetCoins(), sdk.Coins{sdk.NewInt64Coin("stake", 30000000)})
	require.Equal(t, sdk.Coins(nil), vva.DebtAfterFailedVesting)
	// require that the supply has decreased by period 1 amount
	require.Equal(t, initialSupply.Sub(vva.VestingPeriods[0].Amount), supplyKeeper.GetSupply(ctx).GetTotal())
}
