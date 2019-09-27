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
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/keeper"
)

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

	height := int64(0)
	blockTime := now

	addHour := func(t time.Time) time.Time { return t.Add(1 * time.Hour) }

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

	BeginBlocker(ctx, req, vvk)
	height++
	blockTime = addHour(blockTime)
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, []int64{0, 1}, vva.MissingSignCount)

	header = abci.Header{Height: height, Time: addHour(blockTime)}

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
	require.Equal(t, []int64{1, 2}, vva.MissingSignCount)

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
	height++
	blockTime = addHour(blockTime)
	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	require.Equal(t, []int64{2, 3}, vva.MissingSignCount)
}

func TestBeginBlockerSuccessfulPeriod(t *testing.T) {
	height := int64(0)
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

		if height == 11 {
			vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
			// require that missing sign count is set back to zero after the period increments.
			require.Equal(t, []int64{0, 0}, vva.MissingSignCount)
		}

	}

	vva = vvk.GetAccountFromAuthKeeper(ctx, vva.Address)
	// t.Log(vva.MarshalYAML())
	require.Equal(t, [][]int{[]int{1, 1}, []int{0, 0}, []int{0, 0}}, vva.VestingPeriodProgress)
}

func TestBeginBlockerUnsuccessfulPeriod(t *testing.T) {
	height := int64(0)
	now := tmtime.Now()
	blockTime := now
	numBlocks := int64(12)
	addHour := func(t time.Time) time.Time { return t.Add(1 * time.Hour) }

	ctx, ak, _, stakingKeeper, _, vvk := keeper.CreateTestInput(t, false, 1000)
	keeper.CreateValidators(ctx, stakingKeeper, []int64{5, 5, 5})

	vva := keeper.ValidatorVestingDelegatorTestAccount(now)

	ak.SetAccount(ctx, vva)
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
	// check that the period was unsucessful
	require.Equal(t, [][]int{[]int{1, 0}, []int{0, 0}, []int{0, 0}}, vva.VestingPeriodProgress)
	// check that there is debt after the period.
	require.Equal(t, sdk.Coins{sdk.NewInt64Coin("stake", 30000000)}, vva.DebtAfterFailedVesting)

	var delegations int
	stakingKeeper.IterateDelegations(ctx, vva.Address, func(index int64, d stakingexported.DelegationI) (stop bool) {
		delegations++
		return false
	})
	// require that all delegations were unbonded
	require.Equal(t, 0, delegations)
}
