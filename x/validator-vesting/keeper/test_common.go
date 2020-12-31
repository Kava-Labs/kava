package keeper

// nolint:deadcode unused
// DONTCOVER
// noalias
import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/kava-labs/kava/x/bank"
	"github.com/kava-labs/kava/x/validator-vesting/types"
)

//nolint: deadcode unused
var (
	delPk1   = ed25519.GenPrivKey().PubKey()
	delPk2   = ed25519.GenPrivKey().PubKey()
	delPk3   = ed25519.GenPrivKey().PubKey()
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delAddr2 = sdk.AccAddress(delPk2.Address())
	delAddr3 = sdk.AccAddress(delPk3.Address())

	ValOpPk1    = ed25519.GenPrivKey().PubKey()
	ValOpPk2    = ed25519.GenPrivKey().PubKey()
	ValOpPk3    = ed25519.GenPrivKey().PubKey()
	ValOpAddr1  = sdk.ValAddress(ValOpPk1.Address())
	ValOpAddr2  = sdk.ValAddress(ValOpPk2.Address())
	ValOpAddr3  = sdk.ValAddress(ValOpPk3.Address())
	valAccAddr1 = sdk.AccAddress(ValOpPk1.Address()) // generate acc addresses for these validator keys too
	valAccAddr2 = sdk.AccAddress(ValOpPk2.Address())
	valAccAddr3 = sdk.AccAddress(ValOpPk3.Address())

	ValConsPk11  = ed25519.GenPrivKey().PubKey()
	ValConsPk12  = ed25519.GenPrivKey().PubKey()
	ValConsPk13  = ed25519.GenPrivKey().PubKey()
	ValConsAddr1 = sdk.ConsAddress(ValConsPk11.Address())
	ValConsAddr2 = sdk.ConsAddress(ValConsPk12.Address())
	ValConsAddr3 = sdk.ConsAddress(ValConsPk13.Address())

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

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)

	stakingParams := staking.NewParams(time.Hour, 100, uint16(7), 0, sdk.DefaultBondDenom)

	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), blacklistedAddrs)
	maccPerms := map[string][]string{
		auth.FeeCollectorName:     nil,
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		types.ModuleName:          {supply.Burner},
	}
	supplyKeeper := supply.NewKeeper(cdc, keySupply, accountKeeper, bankKeeper, maccPerms)

	stakingKeeper := staking.NewKeeper(cdc, keyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace))
	stakingKeeper.SetParams(ctx, stakingParams)

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
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}

	testAddr := types.CreateTestAddrs(1)[0]
	testPk := types.CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	err := bacc.SetCoins(origCoins)
	if err != nil {
		panic(err)
	}
	vva := types.NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
	err = vva.Validate()
	if err != nil {
		panic(err)
	}
	return vva
}

func ValidatorVestingTestAccounts(numAccounts int) []*types.ValidatorVestingAccount {
	now := tmtime.Now()
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 500), sdk.NewInt64Coin(stakeDenom, 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(feeDenom, 250), sdk.NewInt64Coin(stakeDenom, 25)}},
	}
	testAddr := types.CreateTestAddrs(numAccounts)
	testPk := types.CreateTestPubKeys(numAccounts)
	var vvas []*types.ValidatorVestingAccount
	for i := 0; i < numAccounts; i++ {

		testConsAddr := sdk.ConsAddress(testPk[i].Address())
		origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
		bacc := auth.NewBaseAccountWithAddress(testAddr[i])
		err := bacc.SetCoins(origCoins)
		if err != nil {
			panic(err)
		}
		vva := types.NewValidatorVestingAccount(&bacc, now.Unix(), periods, testConsAddr, nil, 90)
		err = vva.Validate()
		if err != nil {
			panic(err)
		}
		vvas = append(vvas, vva)
	}
	return vvas
}

func ValidatorVestingDelegatorTestAccount(startTime time.Time) *types.ValidatorVestingAccount {
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(stakeDenom, 30000000)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(stakeDenom, 15000000)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin(stakeDenom, 15000000)}},
	}
	testAddr := types.CreateTestAddrs(1)[0]
	testPk := types.CreateTestPubKeys(1)[0]
	testConsAddr := sdk.ConsAddress(testPk.Address())
	origCoins := sdk.Coins{sdk.NewInt64Coin(stakeDenom, 60000000)}
	bacc := auth.NewBaseAccountWithAddress(testAddr)
	err := bacc.SetCoins(origCoins)
	if err != nil {
		panic(err)
	}
	vva := types.NewValidatorVestingAccount(&bacc, startTime.Unix(), periods, testConsAddr, nil, 90)
	err = vva.Validate()
	if err != nil {
		panic(err)
	}
	return vva
}

func CreateValidators(ctx sdk.Context, sk staking.Keeper, powers []int64) {
	val1 := staking.NewValidator(ValOpAddr1, ValOpPk1, staking.Description{})
	val2 := staking.NewValidator(ValOpAddr2, ValOpPk2, staking.Description{})
	val3 := staking.NewValidator(ValOpAddr3, ValOpPk3, staking.Description{})

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
