package validatorvesting

// nolint
// DONTCOVER

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/types"
)

var (
	valTokens  = sdk.TokensFromConsensusPower(42)
	initTokens = sdk.TokensFromConsensusPower(100000)
	valCoins   = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens))
	initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

type testInput struct {
	mApp     *mock.App
	keeper   keeper.Keeper
	sk       staking.Keeper
	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

func getMockApp(t *testing.T, numGenAccs int, genState types.GenesisState, genAccs []auth.Account) testInput {
	mApp := mock.NewApp()

	staking.RegisterCodec(mApp.Cdc)
	types.RegisterCodec(mApp.Cdc)
	supply.RegisterCodec(mApp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	keyValidatorVesting := sdk.NewKVStoreKey(types.StoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)

	validatorVestingAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Burner)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[validatorVestingAcc.GetAddress().String()] = true
	blacklistedAddrs[notBondedPool.GetAddress().String()] = true
	blacklistedAddrs[bondPool.GetAddress().String()] = true

	pk := mApp.ParamsKeeper

	bk := bank.NewBaseKeeper(mApp.AccountKeeper, mApp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)

	maccPerms := map[string][]string{
		types.ModuleName:          {supply.Burner},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
	}
	supplyKeeper := supply.NewKeeper(mApp.Cdc, keySupply, mApp.AccountKeeper, bk, maccPerms)
	sk := staking.NewKeeper(
		mApp.Cdc, keyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace,
	)

	keeper := keeper.NewKeeper(
		mApp.Cdc, keyValidatorVesting, mApp.AccountKeeper, bk, supplyKeeper, sk)

	mApp.SetBeginBlocker(getBeginBlocker(keeper))
	mApp.SetInitChainer(getInitChainer(mApp, keeper, sk, supplyKeeper, genAccs, genState,
		[]supplyexported.ModuleAccountI{validatorVestingAcc, notBondedPool, bondPool}))

	require.NoError(t, mApp.CompleteSetup(keyStaking, keyValidatorVesting, keySupply))

	var (
		addrs    []sdk.AccAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs, valCoins)
	}

	mock.SetGenesis(mApp, genAccs)

	return testInput{mApp, keeper, sk, addrs, pubKeys, privKeys}
}

// gov and staking endblocker
func getBeginBlocker(keeper Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		BeginBlocker(ctx, req, keeper)
		return abci.ResponseBeginBlock{}
	}
}

// gov and staking initchainer
func getInitChainer(mapp *mock.App, keeper Keeper, stakingKeeper staking.Keeper, supplyKeeper supply.Keeper, accs []auth.Account, genState GenesisState,
	blacklistedAddrs []supplyexported.ModuleAccountI) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		stakingGenesis := staking.DefaultGenesisState()

		totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens.MulRaw(int64(len(mapp.GenesisAccounts)))))
		supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

		// set module accounts
		for _, macc := range blacklistedAddrs {
			supplyKeeper.SetModuleAccount(ctx, macc)
		}

		validators := staking.InitGenesis(ctx, stakingKeeper, mapp.AccountKeeper, supplyKeeper, stakingGenesis)
		if genState.IsEmpty() {
			InitGenesis(ctx, keeper, types.DefaultGenesisState())
		} else {
			InitGenesis(ctx, keeper, genState)
		}
		return abci.ResponseInitChain{
			Validators: validators,
		}
	}
}
