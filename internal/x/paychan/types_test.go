package paychan

import (
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
