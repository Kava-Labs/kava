package v3_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	v3 "github.com/kava-labs/kava/x/evmutil/migrations/v3"
)

func getAccountCallback(ctx sdk.Context, addr common.Address) {}

func mkSetAccountCallback(t *testing.T) func(ctx sdk.Context, addr common.Address, account statedb.Account) error {
	return func(ctx sdk.Context, addr common.Address, account statedb.Account) error {
		require.True(t, account.Nonce >= 0 && account.Nonce <= 1)
		return nil
	}
}

func TestMigratePrecompiles(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("evmutil")
	ctx := testutil.DefaultContext(storeKey, sdk.NewTransientStoreKey("transient_test"))

	ctrl := gomock.NewController(t)
	evmKeeperMock := NewMockEvmKeeper(ctrl)
	evmKeeperMock.EXPECT().GetAccount(gomock.Any(), v3.ContractAddress).Do(getAccountCallback).Times(2)
	evmKeeperMock.EXPECT().SetAccount(gomock.Any(), v3.ContractAddress, gomock.Any()).DoAndReturn(mkSetAccountCallback(t)).Times(2)

	evmKeeperMock.EXPECT().SetCode(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)

	err := v3.Migrate(ctx, evmKeeperMock)
	require.NoError(t, err)
}
