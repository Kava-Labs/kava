package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/kava/app/ante"
)

func TestActivateAfterDecorator_AnteHandle(t *testing.T) {

	var upgradeHeight int64 = 1000

	testCases := []struct {
		name   string
		height int64
	}{
		{
			name:   "zero height",
			height: 0,
		},
		{
			name:   "before upgrade height wrapped decorator doesn't run",
			height: upgradeHeight - 1,
		},
		{
			name:   "at upgrade height wrapped decorator runs",
			height: upgradeHeight,
		},
		{
			name:   "after upgrade height wrapped decorator runs",
			height: upgradeHeight + 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrappedDecorator := &MockAnteDecorator{}
			decorator := ante.ActivateAfter(wrappedDecorator, upgradeHeight)
			mah := MockAnteHandler{}

			ctx := sdk.Context{}.WithBlockHeight(tc.height)

			_, err := decorator.AnteHandle(ctx, nil, false, mah.AnteHandle)
			require.NoError(t, err)

			require.True(t, mah.WasCalled)
			shouldHaveRan := tc.height >= upgradeHeight
			require.Equal(t, shouldHaveRan, wrappedDecorator.WasCalled)
		})
	}
}
