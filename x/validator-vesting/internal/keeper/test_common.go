package keeper

// nolint:deadcode unused
// DONTCOVER
// noalias
import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/types/time"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/cosmos/cosmos-sdk/x/validator-vesting/internal/types"
)

//nolint: deadcode unused
var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delPk2   = ed25519.GenPrivKey().PubKey()
	delPk3   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delAddr2 = sdk.AccAddress(delPk2.Address())
	delAddr3 = sdk.AccAddress(delPk3.Address())

	valOpPk1    = ed25519.GenPrivKey().PubKey()
	valOpPk2    = ed25519.GenPrivKey().PubKey()
	valOpPk3    = ed25519.GenPrivKey().PubKey()
	valOpAddr1  = sdk.ValAddress(valOpPk1.Address())
	valOpAddr2  = sdk.ValAddress(valOpPk2.Address())
	valOpAddr3  = sdk.ValAddress(valOpPk3.Address())
	valAccAddr1 = sdk.AccAddress(valOpPk1.Address()) // generate acc addresses for these validator keys too
	valAccAddr2 = sdk.AccAddress(valOpPk2.Address())
	valAccAddr3 = sdk.AccAddress(valOpPk3.Address())

	valConsPk1   = ed25519.GenPrivKey().PubKey()
	valConsPk2   = ed25519.GenPrivKey().PubKey()
	valConsPk3   = ed25519.GenPrivKey().PubKey()
	valConsAddr1 = sdk.ConsAddress(valConsPk1.Address())
	valConsAddr2 = sdk.ConsAddress(valConsPk2.Address())
	valConsAddr3 = sdk.ConsAddress(valConsPk3.Address())

	// TODO move to common testing package for all modules
	// test addresses
	TestAddrs = []sdk.AccAddress{
		delAddr1, delAddr2, delAddr3,
		valAccAddr1, valAccAddr2, valAccAddr3,
	}

	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
	emptyPubkey  crypto.PubKey
	stakeDenom   = "stake"
	feeDenom     = "fee"
)

func MakeTestCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	vesting.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

// test common should produce a staking keeper, a supply keeper, a bank keeper, an auth keeper, a validatorvesting keeper, a context,

func CreateTestInput(t *testing.T, isCheckTx bool, initPower int64) (sdk.Context, auth.AccountKeeper, bank.Keeper, staking.Keeper, supply.Keeper, Keeper) {

	initTokens := sdk.TokensFromConsensusPower(initPower)

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)
	keyValidatorVesting := sdk.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyValidatorVesting, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	require.Nil(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "foo-chain"}, isCheckTx, log.NewNopLogger())

	feeCollectorAcc := supply.NewEmptyModuleAccount(auth.FeeCollectorName)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)
	validatorVestingAcc := supply.NewEmptyModuleAccount(types.ModuleName)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.GetAddress().String()] = true
	blacklistedAddrs[notBondedPool.GetAddress().String()] = true
	blacklistedAddrs[bondPool.GetAddress().String()] = true
	blacklistedAddrs[validatorVestingAcc.GetAddress().String()] = true

	cdc := MakeTestCodec()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)

	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), bank.DefaultCodespace, blacklistedAddrs)
	maccPerms := map[string][]string{
		auth.FeeCollectorName:     nil,
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		types.ModuleName:          {supply.Burner},
	}
	supplyKeeper := supply.NewKeeper(cdc, keySupply, accountKeeper, bankKeeper, maccPerms)

	stakingKeeper := staking.NewKeeper(cdc, keyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	stakingKeeper.SetParams(ctx, staking.DefaultParams())

	keeper := NewKeeper(cdc, keyValidatorVesting, accountKeeper, bankKeeper, supplyKeeper, stakingKeeper)

	initCoins := sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), initTokens))
	totalSupply := sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), initTokens.MulRaw(int64(len(TestAddrs)))))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	// fill all the addresses with some coins, set the loose pool tokens simultaneously
	for _, addr := range TestAddrs {
		_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
		require.Nil(t, err)
	}

	// set module accounts
	keeper.supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	keeper.supplyKeeper.SetModuleAccount(ctx, notBondedPool)
	keeper.supplyKeeper.SetModuleAccount(ctx, bondPool)

	return ctx, accountKeeper, bankKeeper, stakingKeeper, supplyKeeper, keeper
}

func ValidatorVestingTestAccount() *types.ValidatorVestingAccount {
	now := tmtime.Now()
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := types.CreateTestAddrs(1)[0]
	testPk := types.CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	vva := types.NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
	err := vva.Validate()
	if err != nil {
		panic(err)
	}
	return vva
}

func ValidatorVestingDelegatorTestAccount(startTime time.Time) *types.ValidatorVestingAccount {
	periods := vesting.VestingPeriods{
		vesting.VestingPeriod{PeriodLength: int64(12 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(stakeDenom, 30000000)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(stakeDenom, 15000000)}},
		vesting.VestingPeriod{PeriodLength: int64(6 * 60 * 60), VestingAmount: sdk.Coins{sdk.NewInt64Coin(stakeDenom, 15000000)}},
	}
	testAddr := types.CreateTestAddrs(1)[0]
	testPk := types.CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(stakeDenom, 60000000)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	bacc.SetCoins(origCoins)
	vva := types.NewValidatorVestingAccount(&bacc, startTime.Unix(), periods, testConsAddr, nil, 90)
	err := vva.Validate()
	if err != nil {
		panic(err)
	}
	return vva
}

func CreateValidators(ctx sdk.Context, sk staking.Keeper, powers []int64) {
	val1 := staking.NewValidator(valOpAddr1, valOpPk1, staking.Description{})
	val2 := staking.NewValidator(valOpAddr2, valOpPk2, staking.Description{})
	val3 := staking.NewValidator(valOpAddr3, valOpPk3, staking.Description{})

	sk.SetValidator(ctx, val1)
	sk.SetValidator(ctx, val2)
	sk.SetValidator(ctx, val3)
	sk.SetValidatorByConsAddr(ctx, val1)
	sk.SetValidatorByConsAddr(ctx, val2)
	sk.SetValidatorByConsAddr(ctx, val3)
	sk.SetNewValidatorByPowerIndex(ctx, val1)
	sk.SetNewValidatorByPowerIndex(ctx, val2)
	sk.SetNewValidatorByPowerIndex(ctx, val3)

	_, _ = sk.Delegate(ctx, valAccAddr1, sdk.TokensFromConsensusPower(powers[0]), sdk.Unbonded, val1, true)
	_, _ = sk.Delegate(ctx, valAccAddr2, sdk.TokensFromConsensusPower(powers[1]), sdk.Unbonded, val2, true)
	_, _ = sk.Delegate(ctx, valAccAddr3, sdk.TokensFromConsensusPower(powers[2]), sdk.Unbonded, val3, true)

	_ = staking.EndBlocker(ctx, sk)
}
