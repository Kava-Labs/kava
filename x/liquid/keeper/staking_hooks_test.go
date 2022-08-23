package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/liquid/keeper"
	"github.com/kava-labs/kava/x/liquid/types/mocks"
)

func TestStakingHooksTestSuite(t *testing.T) {
	suite.Run(t, new(StakingHooksTestSuite))
}

type StakingHooksTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	keeper        *keeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
}

func (suite *StakingHooksTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	cdc := tApp.AppCodec()

	stakingKey := sdk.NewKVStoreKey(stakingtypes.StoreKey)
	stakingKeeper := stakingkeeper.NewKeeper(cdc, stakingKey, dummyStakingAccountKeeper{}, nil, createDummyParamSubspace())
	suite.stakingKeeper = &stakingKeeper

	keeper := keeper.NewDefaultKeeper(cdc, nil, createDummyParamSubspace(), nil, nil, suite.stakingKeeper)
	suite.keeper = &keeper

	suite.ctx = NewTestContext(stakingKey)
}

// DelegationSharesEqual checks if a delegation has the specified shares.
// It expects delegations with zero shares to not be stored in state.
func (suite *StakingHooksTestSuite) DelegationSharesEqual(valAddr sdk.ValAddress, delegator sdk.AccAddress, shares sdk.Dec) bool {
	del, found := suite.stakingKeeper.GetDelegation(suite.ctx, delegator, valAddr)

	if shares.IsZero() {
		return suite.Falsef(found, "expected delegator to not be found, got %s shares", del.Shares)
	} else {
		res := suite.True(found, "expected delegator to be found")
		return res && suite.Truef(shares.Equal(del.Shares), "expected %s delegator shares but got %s", shares, del.Shares)
	}
}

// NewTestContext sets up a basic context with an in-memory db
func NewTestContext(requiredStoreKeys ...sdk.StoreKey) sdk.Context {
	memDB := db.NewMemDB()
	cms := store.NewCommitMultiStore(memDB)

	for _, key := range requiredStoreKeys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	if err := cms.LoadLatestVersion(); err != nil {
		panic(err)
	}

	return sdk.NewContext(cms, tmprototypes.Header{}, false, log.NewNopLogger())
}

func (suite *StakingHooksTestSuite) TestTransferDelegation_Hooks() {
	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	valAccAddr, fromDelegator, toDelegator := addrs[0], addrs[1], addrs[2]
	valAddr := sdk.ValAddress(valAccAddr)

	testCases := []struct {
		name              string
		transferAllShares bool
		receivingExists   bool
	}{
		{
			name:              "transfer some shares to new delegation",
			transferAllShares: false,
			receivingExists:   false,
		},
		{
			name:              "transfer some shares to existing delegation",
			transferAllShares: false,
			receivingExists:   true,
		},
		{
			name:              "transfer all shares to new delegation",
			transferAllShares: true,
			receivingExists:   false,
		},
		{
			name:              "transfer all shares to existing delegation",
			transferAllShares: true,
			receivingExists:   true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			hooks := mocks.NewStakingHooks(suite.T())
			suite.stakingKeeper.SetHooks(hooks)

			fromShares := d("1000000000.0")
			suite.stakingKeeper.SetValidator(suite.ctx, stakingtypes.Validator{OperatorAddress: valAddr.String()})
			suite.stakingKeeper.SetDelegation(suite.ctx, stakingtypes.NewDelegation(fromDelegator, valAddr, fromShares))

			toShares := sdk.ZeroDec()
			if tc.receivingExists {
				toShares = d("1000000.0")
				suite.stakingKeeper.SetDelegation(suite.ctx, stakingtypes.NewDelegation(toDelegator, valAddr, toShares))
			}

			transferShares := fromShares
			if !tc.transferAllShares {
				transferShares = transferShares.Sub(sdk.OneDec())
			}

			checkDelegationsUnchanged := func(_ mock.Arguments) {
				suite.DelegationSharesEqual(valAddr, fromDelegator, fromShares)
				suite.DelegationSharesEqual(valAddr, toDelegator, toShares)
			}

			// Check both delegations have been updated when either hook is called.
			checkDelegationsUpdated := func(_ mock.Arguments) {
				suite.DelegationSharesEqual(valAddr, fromDelegator, fromShares.Sub(transferShares))
				suite.DelegationSharesEqual(valAddr, toDelegator, toShares.Add(transferShares))
			}

			// from
			hooks.On("BeforeDelegationSharesModified", suite.ctx, fromDelegator, valAddr).
				Run(checkDelegationsUnchanged).Once()
			if tc.transferAllShares {
				hooks.On("BeforeDelegationRemoved", suite.ctx, fromDelegator, valAddr).
					Run(checkDelegationsUnchanged).Once()
			} else {
				hooks.On("AfterDelegationModified", suite.ctx, fromDelegator, valAddr).
					Run(checkDelegationsUpdated).Once()
			}

			// to
			if tc.receivingExists {
				hooks.On("BeforeDelegationSharesModified", suite.ctx, toDelegator, valAddr).
					Run(checkDelegationsUnchanged).Once()
			} else {
				hooks.On("BeforeDelegationCreated", suite.ctx, toDelegator, valAddr).
					Run(checkDelegationsUnchanged).Once()
			}
			hooks.On("AfterDelegationModified", suite.ctx, toDelegator, valAddr).
				Run(checkDelegationsUpdated).Once()

			err := suite.keeper.TransferDelegation(suite.ctx, valAddr, fromDelegator, toDelegator, transferShares)
			suite.NoError(err)
		})
	}
}

// createDummySubspace creates a non functional subspace to satisfy NewKeeper methods.
func createDummyParamSubspace() paramtypes.Subspace {
	return paramtypes.NewSubspace(nil, nil, nil, nil, "")
}

// dummyStakingAccountKeeper is a non functional Keeper that satisfies the staking NewKeeper method.
type dummyStakingAccountKeeper struct{}

func (dummyStakingAccountKeeper) IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) (stop bool)) {
}
func (dummyStakingAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	return nil
}
func (dummyStakingAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress("not nil")
}
func (dummyStakingAccountKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI {
	return nil
}
func (dummyStakingAccountKeeper) SetModuleAccount(sdk.Context, authtypes.ModuleAccountI) {}
