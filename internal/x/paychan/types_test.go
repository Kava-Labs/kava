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
		p := Payout{sdk.Coins{sdk.NewCoin("USD", 4), sdk.NewCoin("GBP", 0)}, sdk.Coins{sdk.NewCoin("USD", 129879234), sdk.NewCoin("GBP", 1)}}
		assert.True(t, p.IsNotNegative())

		p = Payout{sdk.Coins{sdk.NewCoin("USD", -4), sdk.NewCoin("GBP", 0)}, sdk.Coins{sdk.NewCoin("USD", 129879234), sdk.NewCoin("GBP", 1)}}
		assert.False(t, p.IsNotNegative())
	})
	t.Run("Sum", func(t *testing.T) {
		p := Payout{
			sdk.Coins{sdk.NewCoin("EUR", 1), sdk.NewCoin("USD", -5)},
			sdk.Coins{sdk.NewCoin("EUR", 1), sdk.NewCoin("USD", 100), sdk.NewCoin("GBP", 1)},
		}
		expected := sdk.Coins{sdk.NewCoin("EUR", 2), sdk.NewCoin("GBP", 1), sdk.NewCoin("USD", 95)}
		assert.Equal(t, expected, p.Sum())
	})
}
