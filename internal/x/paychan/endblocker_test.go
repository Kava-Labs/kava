package paychan

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestEndBlocker(t *testing.T) {
	// TODO test that endBlocker doesn't close channels before the execution time

	// SETUP
	accountSeeds := []string{"senderSeed", "receiverSeed"}
	ctx, _, channelKeeper, addrs, _, _, _ := createMockApp(accountSeeds)
	sender := addrs[0]
	receiver := addrs[1]
	coins := sdk.Coins{sdk.NewInt64Coin("KVA", 10)}

	// create new channel
	channelID := ChannelID(0) // should be 0 as first channel
	channel := Channel{
		ID:           channelID,
		Participants: [2]sdk.AccAddress{sender, receiver},
		Coins:        coins,
	}
	channelKeeper.setChannel(ctx, channel)

	// create closing update and submittedUpdate
	payout := Payout{sdk.Coins{sdk.NewInt64Coin("KVA", 3)}, sdk.Coins{sdk.NewInt64Coin("KVA", 7)}}
	update := Update{
		ChannelID: channelID,
		Payout:    payout,
	}
	sUpdate := SubmittedUpdate{
		Update:        update,
		ExecutionTime: 0, // current blocktime
	}
	// Set empty submittedUpdatesQueue TODO work out proper genesis initialisation
	channelKeeper.setSubmittedUpdatesQueue(ctx, SubmittedUpdatesQueue{})
	// flag channel for closure
	channelKeeper.addToSubmittedUpdatesQueue(ctx, sUpdate)

	// ACTION
	EndBlocker(ctx, channelKeeper)

	// CHECK RESULTS
	// ideally just check if keeper.channelClose was called, but can't
	// writing endBlocker to accept an interface of which keeper is implementation would make this possible
	// check channel is gone
	_, found := channelKeeper.getChannel(ctx, channelID)
	assert.False(t, found)
	// check queue is empty, NOTE: due to encoding, an empty queue (underneath just an int slice) will be decoded as nil slice rather than an empty slice
	suq := channelKeeper.getSubmittedUpdatesQueue(ctx)
	assert.Equal(t, SubmittedUpdatesQueue(nil), suq)
	// check submittedUpdate is gone
	_, found = channelKeeper.getSubmittedUpdate(ctx, channelID)
	assert.False(t, found)
}
