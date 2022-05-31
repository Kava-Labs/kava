package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/ante"
)

func TestActivateAfterDecorator_AnteHandle(t *testing.T) {

	tApp := app.NewTestApp()
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
			mmd := MockAnteHandler{}

			ctx := tApp.NewContext(true, tmproto.Header{Height: tc.height, Time: tmtime.Now()})

			_, err := decorator.AnteHandle(ctx, nil, false, mmd.AnteHandle)
			require.NoError(t, err)

			require.True(t, mmd.WasCalled)
			shouldHaveRan := tc.height >= upgradeHeight
			require.Equal(t, shouldHaveRan, wrappedDecorator.WasCalled)
		})
	}
}
