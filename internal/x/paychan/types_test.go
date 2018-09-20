package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubmittedUpdatesQueue(t *testing.T) {
	t.Run("RemoveMatchingElements", func(t *testing.T) {
		// SETUP
		q := SubmittedUpdatesQueue{4, 8, 23, 0, 5645657}
		// ACTION
		q.RemoveMatchingElements(23)
		// CHECK RESULTS
		expectedQ := SubmittedUpdatesQueue{4, 8, 0, 5645657}
		assert.Equal(t, expectedQ, q)

		// SETUP
		q = SubmittedUpdatesQueue{0}
		// ACTION
		q.RemoveMatchingElements(0)
		// CHECK RESULTS
		expectedQ = SubmittedUpdatesQueue{}
		assert.Equal(t, expectedQ, q)
	})
}

func TestPayout(t *testing.T) {
	t.Run("IsNotNegative", func(t *testing.T) {
		p := Payout{sdk.Coins{sdk.NewInt64Coin("USD", 4), sdk.NewInt64Coin("GBP", 0)}, sdk.Coins{sdk.NewInt64Coin("USD", 129879234), sdk.NewInt64Coin("GBP", 1)}}
		assert.True(t, p.IsNotNegative())

		p = Payout{sdk.Coins{sdk.NewInt64Coin("USD", -4), sdk.NewInt64Coin("GBP", 0)}, sdk.Coins{sdk.NewInt64Coin("USD", 129879234), sdk.NewInt64Coin("GBP", 1)}}
		assert.False(t, p.IsNotNegative())
	})
	t.Run("Sum", func(t *testing.T) {
		p := Payout{
			sdk.Coins{sdk.NewInt64Coin("EUR", 1), sdk.NewInt64Coin("USD", -5)},
			sdk.Coins{sdk.NewInt64Coin("EUR", 1), sdk.NewInt64Coin("USD", 100), sdk.NewInt64Coin("GBP", 1)},
		}
		expected := sdk.Coins{sdk.NewInt64Coin("EUR", 2), sdk.NewInt64Coin("GBP", 1), sdk.NewInt64Coin("USD", 95)}
		assert.Equal(t, expected, p.Sum())
	})
}
