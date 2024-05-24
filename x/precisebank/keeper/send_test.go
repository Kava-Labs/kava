package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/kava-labs/kava/x/precisebank/types"
	"github.com/stretchr/testify/require"
)

func TestSendCoinsFromAccountToModule_BlockedReserve(t *testing.T) {
	// Other modules shouldn't be able to send x/precisebank coins as the module
	// account balance is for internal reserve use only.

	td := NewMockedTestData(t)
	td.ak.EXPECT().
		GetModuleAccount(td.ctx, types.ModuleName).
		Return(authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(sdk.AccAddress{100}),
			types.ModuleName,
		)).
		Once()

	fromAddr := sdk.AccAddress([]byte{1})
	err := td.keeper.SendCoinsFromAccountToModule(td.ctx, fromAddr, types.ModuleName, cs(c("busd", 1000)))

	require.Error(t, err)
	require.EqualError(t, err, "module account precisebank is not allowed to receive funds: unauthorized")
}

func TestSendCoinsFromModuleToAccount_BlockedReserve(t *testing.T) {
	// Other modules shouldn't be able to send x/precisebank module account
	// funds.

	td := NewMockedTestData(t)
	td.ak.EXPECT().
		GetModuleAddress(types.ModuleName).
		Return(sdk.AccAddress{100}).
		Once()

	toAddr := sdk.AccAddress([]byte{1})
	err := td.keeper.SendCoinsFromModuleToAccount(td.ctx, types.ModuleName, toAddr, cs(c("busd", 1000)))

	require.Error(t, err)
	require.EqualError(t, err, "module account precisebank is not allowed to send funds: unauthorized")
}
